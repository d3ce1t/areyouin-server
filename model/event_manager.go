package model

import (
	"bytes"
	"crypto/sha256"
	"image"
	"log"
	"peeple/areyouin/api"
	"peeple/areyouin/cqldao"
	"peeple/areyouin/utils"
	"time"

	"github.com/imkira/go-observer"
)

type EventManager struct {
	dbsession       api.DbSession
	parent          *AyiModel
	userDAO         api.UserDAO
	eventDAO        api.EventDAO
	timelineDAO     api.EventTimeLineDAO
	eventHistoryDAO api.EventHistoryDAO
	thumbDAO        api.ThumbnailDAO
	eventSignal     observer.Property
	userEvents      *UserEvents
}

func newEventManager(parent *AyiModel, session api.DbSession) *EventManager {
	evManager := &EventManager{
		parent:          parent,
		dbsession:       session,
		userDAO:         cqldao.NewUserDAO(session),
		eventDAO:        cqldao.NewEventDAO(session),
		eventHistoryDAO: cqldao.NewEventHistoryDAO(session),
		timelineDAO:     cqldao.NewTimeLineDAO(session),
		thumbDAO:        cqldao.NewThumbnailDAO(session),
		eventSignal:     observer.NewProperty(nil),
		userEvents:      newUserEvents(),
	}

	if err := evManager.loadActiveEvents(); err != nil {
		log.Printf("newEventManagerError: %v\n", err)
		return nil
	}

	return evManager
}

func (m *EventManager) LoadEvent(eventID int64) (*Event, error) {

	events, err := m.eventDAO.LoadEvents(eventID)
	if err != nil {
		return nil, err
	}

	if len(events) > 0 {
		return newEventFromDTO(events[0]), nil
	}

	return nil, ErrNotFound
}

/*
 * Save an event into database in order to let users request his events list
 *
 * Preconditions:
 * (1) event must have been initialised by model, this implies event object
 *     is valid
 * (2) event must not be not persisted
 */
func (m *EventManager) SaveEvent(event *Event) error {

	// Check precondition (1)
	if event == nil || !event.initialised {
		return ErrNotInitialised
	}

	// Check precondition (2)
	if event.isPersisted {
		// Do nothing
		return nil
	}

	if event.oldEvent == nil || !event.oldEvent.isPersisted {

		// New event that has no copy in database
		return m.saveNewEvent(event)

	}

	// Modification of an event that has a copy in database
	return m.saveModifiedEvent(event)
}

// Cancel an existing event
//
// Assumptions:
// - (1) Event exist and is persisted in DB
// - (2) User who performs this operation has permission
//
// Preconditions:
// - (1) Event must have not started
func (m *EventManager) CancelEvent(event *Event, userID int64) error {

	// Check precondition (1)
	if event.Status() != api.EventState_NOT_STARTED {
		return ErrEventNotWritable
	}

	// Change event state and position in time line

	oldPosition := event.EndDate() // already truncated
	newPosition := utils.GetCurrentTimeUTC().Truncate(time.Second)

	err := m.eventDAO.CancelEvent(event.Id(), oldPosition, newPosition, event.ParticipantIds())
	if err != nil {
		return err
	}

	// Update event object
	event.inboxPosition = newPosition
	event.cancelled = true

	// Remove from inbox and add to history
	for pID := range event.participants {
		m.userEvents.Remove(pID, event.Id())
	}

	// Emit signal
	signal := &Signal{
		Type: SignalEventCancelled,
		Data: map[string]interface{}{
			"EventID":     event.Id(),
			"CancelledBy": userID,
			"Event":       event,
		},
	}

	m.eventSignal.Update(signal)

	return nil
}

// Change event picture
//
// Assumptions:
// - (1) Event exist and is persisted in DB
// - (2) User who performs this operation have permission
//
// Preconditions
// - (1) Event must have not started
func (m *EventManager) ChangeEventPicture(event *Event, picture []byte) error {

	// Check precondition (1)
	if event.Status() != api.EventState_NOT_STARTED {
		return ErrEventNotWritable
	}

	modified := false

	if picture != nil && len(picture) != 0 {

		// Set event picture

		// Compute digest for picture
		digest := sha256.Sum256(picture)

		corePicture := &Picture{
			RawData: picture,
			Digest:  digest[:],
		}

		if !bytes.Equal(event.pictureDigest, corePicture.Digest) {

			// Save event picture
			if err := m.saveEventPicture(event.id, corePicture); err != nil {
				return err
			}

			event.pictureDigest = corePicture.Digest
			modified = true
		}

	} else {

		// Remove event picture

		if err := m.removeEventPicture(event.id); err != nil {
			return err
		}

		event.pictureDigest = nil
		modified = true
	}

	// Emit signal
	if modified {
		signal := &Signal{
			Type: SignalEventInfoChanged,
			Data: map[string]interface{}{
				"EventID": event.Id(),
				"Event":   event,
			},
		}

		m.eventSignal.Update(signal)
	}

	return nil
}

// Change participant response to an event.
// Returns true if response changed, or false otherwise. For instance, if response
// is equal to old response, then it would return false.
//
// Assumptions:
// - (1) Event exist and is persisted in DB
// - (2) User who performs this operation have permission
//
// Preconditions:
// - (1) Event must have not started
// - (2) User must have received this invitation, i.e. user is in event participant
//       list and event is in his inbox.
func (m *EventManager) ChangeParticipantResponse(userId int64,
	response api.AttendanceResponse, event *Event) (bool, error) {

	// Check precondition (1)

	if event.Status() != api.EventState_NOT_STARTED {
		return false, ErrEventNotWritable
	}

	// Check precondition (2)

	participant, ok := event.participants[userId]
	if !ok {
		return false, ErrParticipantNotFound
	}

	if participant.response != response {

		// Change response

		if err := m.eventDAO.SetParticipantResponse(participant.id, response, event.AsDTO()); err != nil {
			return false, err
		}

		oldResponse := participant.response
		participant.response = response

		// Emit signal
		signal := &Signal{
			Type: SignalParticipantChanged,
			Data: map[string]interface{}{
				"EventID":     event.Id(),
				"UserID":      participant.Id(),
				"Event":       event,
				"Participant": participant,
				"OldResponse": oldResponse,
			},
		}

		participant.response = response

		m.eventSignal.Update(signal)

		return true, nil
	}

	return false, nil
}

func (m *EventManager) ChangeDeliveryState(event *Event, userId int64, state api.InvitationStatus) (bool, error) {

	if event.Status() != api.EventState_NOT_STARTED {
		return false, ErrEventNotWritable
	}

	participant, ok := event.participants[userId]
	if !ok {
		return false, ErrParticipantNotFound
	}

	if participant.invitationStatus != state {
		// TODO: Add business logic to avoid moving to a previous state

		err := m.eventDAO.SetParticipantInvitationStatus(userId, event.id, state)

		if err != nil {
			return false, err
		}

		participant.invitationStatus = state

		// Emit signal
		signal := &Signal{
			Type: SignalParticipantChanged,
			Data: map[string]interface{}{
				"EventID":     event.Id(),
				"UserID":      participant.Id(),
				"Event":       event,
				"Participant": participant,
				"OldResponse": participant.response,
			},
		}

		m.eventSignal.Update(signal)

		return true, nil
	}

	return false, nil
}

func (m *EventManager) GetEventForUser(userId int64, eventId int64) (*Event, error) {

	// TODO: Make a more efficient implementation by adding DAO support to load a single event

	events, err := m.eventDAO.LoadEvents(eventId)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, ErrNotFound
	}

	event := newEventFromDTO(events[0])

	if participant := event.GetParticipant(userId); participant == nil {
		return nil, ErrNotFound
	}

	return event, nil
}

// FIXME: Do not get all of the private events, but limit to a fixed number.
func (m *EventManager) GetRecentEvents(userID int64) ([]*Event, error) {

	// Event IDs
	eventIDs := m.userEvents.FindAll(userID)

	if len(eventIDs) == 0 {
		return nil, ErrEmptyInbox
	}

	// Read from event table to get the actual info
	eventsDTO, err := m.eventDAO.LoadEvents(eventIDs...)
	if err != nil {
		log.Printf("GetRecentEvents (%v): %v\n", userID, err)
		return nil, err
	}

	// Convert DTO to model.Event
	events := make([]*Event, 0, len(eventsDTO))
	for _, ev := range eventsDTO {
		events = append(events, newEventFromDTO(ev))
	}

	return events, nil
}

func (m *EventManager) GetEventsHistory(userID int64, start time.Time, end time.Time) ([]*Event, error) {

	currentTime := utils.GetCurrentTimeUTC().Truncate(time.Second)

	// Fix start and end to not be higher than current time
	if start.After(currentTime) {
		start = currentTime
	}

	if end.After(currentTime) {
		end = currentTime
	}

	// Load event ID from user history in database
	var eventIDs []int64
	var err error

	if start.Before(end) {
		eventIDs, err = m.eventHistoryDAO.FindAllForward(userID, start.Truncate(time.Second))
	} else {
		eventIDs, err = m.eventHistoryDAO.FindAllBackward(userID, start.Truncate(time.Second))
	}

	if err == api.ErrNoResults {
		return nil, ErrEmptyInbox
	} else if err != nil {
		return nil, err
	}

	// Read from event table to get the actual info
	eventsDTO, err := m.eventDAO.LoadEvents(eventIDs...)
	if err != nil {
		return nil, err
	}

	// Convert DTO to model.Event
	events := make([]*Event, 0, len(eventsDTO))
	for _, ev := range eventsDTO {
		events = append(events, newEventFromDTO(ev))
	}

	return events, nil
}

func (m *EventManager) BuildEventsTimeLine() error {

	if err := m.timelineDAO.DeleteAll(); err != nil {
		return err
	}

	err := m.eventDAO.RangeAll(func(event *api.EventDTO) error {

		// Insert into timeline
		entryDTO := &api.TimeLineEntryDTO{
			EventID:  event.Id,
			Position: utils.MillisToTimeUTC(event.EndDate).Truncate(time.Second),
		}

		if event.Cancelled {
			entryDTO.Position = utils.MillisToTimeUTC(event.InboxPosition).Truncate(time.Second)
		}

		if err := m.timelineDAO.Insert(entryDTO); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (m *EventManager) BuildEventsHistory() error {

	if err := m.eventHistoryDAO.DeleteAll(); err != nil {
		return err
	}

	err := m.eventDAO.RangeAll(func(event *api.EventDTO) error {

		entryDTO := &api.TimeLineEntryDTO{
			EventID:  event.Id,
			Position: utils.MillisToTimeUTC(event.EndDate).Truncate(time.Second),
		}

		if event.Cancelled {
			entryDTO.Position = utils.MillisToTimeUTC(event.InboxPosition).Truncate(time.Second)
		}

		for pID, _ := range event.Participants {

			// Insert into event history
			if err := m.eventHistoryDAO.Insert(pID, entryDTO); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (m *EventManager) Observe() observer.Stream {
	return m.eventSignal.Observe()
}

func (m *EventManager) emitEventoInfoChanged(event *Event) {
	m.eventSignal.Update(&Signal{
		Type: SignalEventInfoChanged,
		Data: map[string]interface{}{
			"EventID": event.Id(),
			"Event":   event,
		},
	})
}

func (m *EventManager) emitEventParticipantsInvited(event *Event, newParticipants []int64, oldParticipants []int64) {
	m.eventSignal.Update(&Signal{
		Type: SignalEventParticipantsInvited,
		Data: map[string]interface{}{
			"EventID":         event.Id(),
			"NewParticipants": newParticipants,
			"OldParticipants": oldParticipants,
			"Event":           event,
		},
	})
}

func (m *EventManager) emitNewEvent(event *Event) {
	m.eventSignal.Update(&Signal{
		Type: SignalNewEvent,
		Data: map[string]interface{}{
			"EventID":         event.Id(),
			"NewParticipants": event.ParticipantIds(),
			"OldParticipants": []int{},
			"Event":           event,
		},
	})
}

/*
 * Preconditions
 * (1) author must exists and be valid
 * (2) event must be created inside a valid temporal window
 */
func (m *EventManager) saveNewEvent(event *Event) error {

	currentDate := utils.GetCurrentTimeUTC().Truncate(time.Second)

	// Check precondition (2)
	if event.CreatedDate().Before(currentDate.Add(-2*time.Minute)) || event.CreatedDate().After(currentDate.Add(2*time.Minute)) {
		return ErrEventOutOfCreationWindow
	}

	// Check precondition (1)
	_, err := m.userDAO.Load(event.AuthorID())
	if err == api.ErrNotFound {
		return ErrInvalidAuthor
	} else if err != nil {
		return err
	}

	// Insert in timeline
	entry := &api.TimeLineEntryDTO{EventID: event.Id(), Position: event.EndDate()}
	if err := m.timelineDAO.Insert(entry); err != nil {
		return err
	}

	// Persist event.
	if err := m.eventDAO.Insert(event.AsDTO()); err != nil {
		return err
	}

	// Add to inbox
	for pID := range event.participants {
		m.userEvents.Insert(pID, event.Id())
	}

	// If code failed before reaching this point, a timeline entry
	// could exist that doesn't point to any event. Moreover, if
	// 'add to inbox' failed, event would exist only in database
	// but users would be unable to retrieve it.

	// Emit signal
	m.emitNewEvent(event)

	return nil
}

/*
 * Preconditions
 * (1) event must be created inside a valid temporal window
 */
func (m *EventManager) saveModifiedEvent(event *Event) error {

	currentDate := utils.GetCurrentTimeUTC().Truncate(time.Second)
	oldEvent := event.oldEvent

	// Check precondition (1)
	if event.ModifiedDate().Before(currentDate.Add(-2*time.Minute)) || event.ModifiedDate().After(currentDate.Add(2*time.Minute)) {
		return ErrEventOutOfCreationWindow
	}

	if !event.endDate.Equal(oldEvent.endDate) {

		// Move event in timeline
		// TODO: Write timeline entry along with the event in a logged batch

		oldEntry := &api.TimeLineEntryDTO{EventID: event.Id(), Position: oldEvent.EndDate()}
		newEntry := &api.TimeLineEntryDTO{EventID: event.Id(), Position: event.EndDate()}

		if err := m.timelineDAO.Replace(oldEntry, newEntry); err != nil {
			return err
		}
	}

	// Persist event. Take into account that only new participants must be written
	// because previous participants could have changed their state since event
	// was loaded
	eventDTO := event.CloneEmptyParticipants().AsDTO()

	newParticipants := m.ExtractNewParticipants(event, oldEvent)
	for pID, participant := range newParticipants {
		eventDTO.Participants[pID] = participant.AsDTO()
	}

	// If this insert fail, timeline entry would be inconsistent
	// TODO: Use transaction (IF bla != known_value) to avoid an event being modified from two places at the same time
	if err := m.eventDAO.Insert(eventDTO); err != nil {
		return err
	}

	// Add to inbox
	for pID := range newParticipants {
		m.userEvents.Insert(pID, event.Id())
	}

	// Emit signal
	if m.isEventInfoChanged(event, oldEvent) {
		m.emitEventoInfoChanged(event)
	}

	// Emit signal
	if len(newParticipants) > 0 {
		m.emitEventParticipantsInvited(event, getParticipantMapKeys(newParticipants), oldEvent.ParticipantIds())
	}

	return nil
}

// ExtractNewParticipants extracts participants from extractList that are not in baseList
func (m *EventManager) ExtractNewParticipants(extractEvent *Event, baseEvent *Event) map[int64]*Participant {

	newParticipants := make(map[int64]*Participant)

	for pID, participant := range extractEvent.participants {
		if _, ok := baseEvent.participants[pID]; !ok {
			newParticipants[pID] = participant
		}
	}

	return newParticipants
}

func (m *EventManager) isEventInfoChanged(event *Event, oldEvent *Event) bool {

	if event.startDate.Equal(oldEvent.startDate) &&
		event.endDate.Equal(oldEvent.endDate) &&
		event.createdDate.Equal(oldEvent.createdDate) &&
		event.inboxPosition.Equal(oldEvent.inboxPosition) &&
		event.description == oldEvent.description &&
		event.cancelled == oldEvent.cancelled {

		return false
	}

	return true
}

func (m *EventManager) saveEventPicture(event_id int64, picture *Picture) error {

	// Decode image
	srcImage, _, err := image.Decode(bytes.NewReader(picture.RawData))
	if err != nil {
		return err
	}

	// Check image size is inside bounds
	if srcImage.Bounds().Dx() > eventPictureMaxWidth || srcImage.Bounds().Dy() > eventPictureMaxHeight {
		return ErrImageOutOfBounds
	}

	// Create thumbnails
	thumbnails, err := utils.CreateThumbnails(srcImage, eventThumbnailSize, m.parent.supportedDpi)
	if err != nil {
		return err
	}

	// Save thumbnails
	err = m.thumbDAO.Insert(event_id, picture.Digest, thumbnails)
	if err != nil {
		return err
	}

	// Save event picture (always does it after thumbnails)
	err = m.eventDAO.SetEventPicture(event_id, picture.AsDTO())
	if err != nil {
		return err
	}

	return nil
}

func (m *EventManager) removeEventPicture(event_id int64) error {

	// Remove event picture
	emptyPic := Picture{RawData: nil, Digest: nil}
	err := m.eventDAO.SetEventPicture(event_id, emptyPic.AsDTO())
	if err != nil {
		return err
	}

	// Remove thumbnails
	err = m.thumbDAO.Remove(event_id)
	if err != nil {
		return err
	}

	return nil
}

// Tells if participant p1 can see event changes of participant p2
func (m *EventManager) canSee(p1 int64, p2 *Participant) (bool, error) {
	if p2.response == api.AttendanceResponse_ASSIST {
		return true, nil
	} else {
		return m.parent.Friends.IsFriend(p2.id, p1)
	}
}

func (m *EventManager) loadActiveEvents() error {

	currentTimeMillis := utils.GetCurrentTimeMillis()

	// Load next events
	entries, err := m.timelineDAO.FindAllFrom(currentTimeMillis)
	if err != nil {
		return err
	} else if len(entries) == 0 {
		return nil
	}

	// Extract IDs
	eventIDs := make([]int64, 0, len(entries))
	for _, entry := range entries {
		eventIDs = append(eventIDs, entry.EventID)
	}

	// Range events
	err = m.eventDAO.RangeEvents(func(event *api.EventDTO) error {

		for pID := range event.Participants {

			if event.Cancelled {
				// Time line should not include cancelled events
				log.Printf("WARNING: Timeline includes cancelled events (currentTime: %v, EventID: %v)\n",
					currentTimeMillis, event.Id)
			}

			// Add entry to each user of the event

			m.userEvents.Insert(pID, event.Id)
		}

		return nil

	}, eventIDs...)

	if err != nil {
		return err
	}

	// Sort user inbox
	/*for _, inbox := range m.userInbox {
		sort.Sort(api.TimeLineByEndDate(inbox))
	}*/

	return nil
}
