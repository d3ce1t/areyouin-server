package model

import (
	"bytes"
	"crypto/sha256"
	"image"
	"log"
	"peeple/areyouin/api"
	"peeple/areyouin/cqldao"
	"peeple/areyouin/utils"
	"strconv"
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
	settingsDAO     api.SettingsDAO
	eventSignal     observer.Property
	userEvents      *UserEvents
	lastArchiveTime time.Time
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
		settingsDAO:     cqldao.NewSettingsDAO(session),
		eventSignal:     observer.NewProperty(nil),
		userEvents:      newUserEvents(),
	}

	if err := evManager.readLastArchiveTime(); err != nil {
		log.Printf("newEventManagerError: %v\n", err)
		panic(ErrModelInitError)
	}

	if err := evManager.loadActiveEvents(); err != nil {
		log.Printf("newEventManagerError: %v\n", err)
		panic(ErrModelInitError)
	}

	return evManager
}

func (m *EventManager) initBackgroundTasks() {
	go func() {
		for {
			m.startJobManager()
		}
	}()
}

func (m *EventManager) readLastArchiveTime() error {

	value, err := m.settingsDAO.Find(api.MasterLastArchiveTime)
	if err == api.ErrNotFound {
		value = "0"
	} else if err != nil {
		return err
	}

	millis, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}

	m.lastArchiveTime = utils.MillisToTimeUTC(millis)
	return nil
}

func (m *EventManager) startJobManager() {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("EventsJobManager: Unexpected end %v", r)
		}
	}()

	log.Println("EventsJobManager: Started")

	var lastTime time.Time

	archiveJob := func() {
		lastTime = utils.GetCurrentTimeUTC()
		if err := m.archiveFinishedEventsSinceLastCheck(); err != nil {
			log.Printf("EventsJobManager: Finish archiving events with errors: %v", err)
		}
	}

	for {
		currentTime := utils.GetCurrentTimeUTC()
		timeSinceLastTime := currentTime.Sub(lastTime)

		if timeSinceLastTime >= 1*time.Minute {
			archiveJob()
		} else {
			nextMinute := currentTime.Truncate(time.Minute).Add(time.Minute)
			select {
			case <-time.After(nextMinute.Sub(currentTime)):
				archiveJob()
			}
		}
	}
}

func (m *EventManager) NewEvent(author *UserAccount, createdDate time.Time, startDate time.Time, endDate time.Time,
	description string, participants []int64) (*Event, error) {

	b := m.NewEventBuilder(author.Id())
	b.SetAuthor(author)
	b.SetCreatedDate(createdDate)
	b.SetStartDate(startDate)
	b.SetEndDate(endDate)
	b.SetDescription(description)

	pAuthor := NewParticipant(author.Id(), author.Name(),
		api.AttendanceResponse_ASSIST, api.InvitationStatus_SERVER_DELIVERED)

	b.ParticipantAdder().AddParticipant(pAuthor)
	for _, pID := range participants {
		b.ParticipantAdder().AddUserID(pID)
	}

	return b.Build()
}

func (m *EventManager) LoadEvent(eventID int64) (*Event, error) {

	events, err := m.eventDAO.LoadEvents(eventID)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, ErrNotFound
	}

	event := newEventFromDTO(events[0])
	event.isPersisted = true

	return event, nil
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
// - (2) User must have received this invitation, i.e. user is in event participant list
func (m *EventManager) ChangeParticipantResponse(userID int64, response api.AttendanceResponse, event *Event) (*Participant, error) {

	// Check precondition (1)

	if event.Status() != api.EventState_NOT_STARTED {
		return nil, ErrEventNotWritable
	}

	// Check precondition (2)

	participant, ok := event.participants[userID]
	if !ok {
		return nil, ErrParticipantNotFound
	}

	if participant.response != response {

		// Change response

		b, err := m.NewParticipantModifier(participant)
		if err != nil {
			return nil, err
		}

		b.SetResponse(response)
		modifiedParticipant, err := b.Build()
		if err != nil {
			return nil, err
		}

		// Persist
		if err := m.eventDAO.InsertParticipant(modifiedParticipant.AsDTO()); err != nil {
			return nil, err
		}

		// Emit signal
		m.emitParticipantChanged(participant, modifiedParticipant, event)

		return modifiedParticipant, nil
	}

	return participant, nil
}

func (m *EventManager) ChangeDeliveryState(event *Event, userId int64, state api.InvitationStatus) (*Participant, error) {

	if event.Status() != api.EventState_NOT_STARTED {
		return nil, ErrEventNotWritable
	}

	participant, ok := event.participants[userId]
	if !ok {
		return nil, ErrParticipantNotFound
	}

	if participant.invitationStatus != state {

		// Change status
		// TODO: Add business logic to avoid moving to a previous state

		b, err := m.NewParticipantModifier(participant)
		if err != nil {
			return nil, err
		}

		b.SetInvitationStatus(state)
		modifiedParticipant, err := b.Build()
		if err != nil {
			return nil, err
		}

		// Persist
		if err := m.eventDAO.InsertParticipant(modifiedParticipant.AsDTO()); err != nil {
			return nil, err
		}

		// Emit signal
		m.emitParticipantChanged(participant, modifiedParticipant, event)

		return modifiedParticipant, nil
	}

	return participant, nil
}

func (m *EventManager) GetEventForUser(userID int64, eventID int64) (*Event, error) {

	event, err := m.LoadEvent(eventID)
	if err != nil {
		return nil, err
	}

	if participant := event.GetParticipant(userID); participant == nil {
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
		event := newEventFromDTO(ev)
		event.isPersisted = true
		events = append(events, event)
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
		event := newEventFromDTO(ev)
		event.isPersisted = true
		events = append(events, event)
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

	currentTime := utils.GetCurrentTimeUTC()

	err := m.eventDAO.RangeAll(func(event *api.EventDTO) error {

		endDate := utils.MillisToTimeUTC(event.EndDate)

		if endDate.After(currentTime) {
			log.Printf("WARNING: Skip archiving event %v because it has not finished yet", event.Id)
			return nil
		}

		if err := m.archiveEvent(event); err != nil {
			return err
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

func (m *EventManager) emitParticipantChanged(oldParticipant *Participant, newParticipant *Participant, event *Event) {
	m.eventSignal.Update(&Signal{
		Type: SignalParticipantChanged,
		Data: map[string]interface{}{
			"EventID":     newParticipant.eventID,
			"UserID":      newParticipant.id,
			"Event":       event,
			"Participant": newParticipant,
			"OldResponse": oldParticipant.response,
		},
	})
}

func (m *EventManager) emitEventCancelled(event *Event, cancelledBy int64) {
	m.eventSignal.Update(&Signal{
		Type: SignalEventCancelled,
		Data: map[string]interface{}{
			"EventID":     event.Id(),
			"CancelledBy": cancelledBy,
			"Event":       event,
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

	// Persist event. Take into account that only new participants must be written
	// because previous participants could have changed their state since event
	// was loaded
	eventDTO := event.CloneEmptyParticipants().AsDTO()
	newParticipants := m.ExtractNewParticipants(event, oldEvent)
	for pID, participant := range newParticipants {
		eventDTO.Participants[pID] = participant.AsDTO()
	}

	// Persist event
	if err := m.eventDAO.Replace(oldEvent.AsDTO(), eventDTO); err != nil {
		return err
	}

	if !event.cancelled {

		// Add this event to user's inbox
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

	} else {

		// Remove this event from user's inbox
		for pID := range event.participants {
			m.userEvents.Remove(pID, event.Id())
		}

		// Emit cancelled
		m.emitEventCancelled(event, event.owner)
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

	currentTime := utils.GetCurrentTimeUTC()

	// Load next events
	entries, err := m.timelineDAO.FindAllForward(currentTime)
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

		if event.Cancelled {
			// Time line should not include cancelled events
			log.Printf("WARNING: Timeline includes cancelled events (currentTime: %v, EventID: %v)\n",
				currentTime, event.Id)
			return nil
		}

		for pID := range event.Participants {

			// Add entry to each user of the event

			m.userEvents.Insert(pID, event.Id)
		}

		return nil

	}, eventIDs...)

	if err != nil {
		return err
	}

	return nil
}

func (m *EventManager) archiveEvent(event *api.EventDTO) error {

	entryDTO := &api.TimeLineEntryDTO{
		EventID:  event.Id,
		Position: utils.MillisToTimeUTC(event.EndDate).Truncate(time.Second),
	}

	if event.Cancelled {
		entryDTO.Position = utils.MillisToTimeUTC(event.InboxPosition).Truncate(time.Second)
	}

	for pID := range event.Participants {

		// Insert into event history
		if err := m.eventHistoryDAO.Insert(pID, entryDTO); err != nil {
			return err
		}
	}

	return nil
}

// Read finished or cancelled events since last time and archives them,
// i.e. events are move from user's recent events to events history
func (m *EventManager) archiveFinishedEventsSinceLastCheck() error {

	currentTime := utils.GetCurrentTimeUTC()

	// Get events between last check and current time, both included.
	// In order to not retrieve events that were loaded in the previous call,
	// add 1 millisecond to the begining of the window.
	from := m.lastArchiveTime.Add(time.Millisecond)
	entries, err := m.timelineDAO.FindAllBetween(from, currentTime)
	if err != nil {
		return err
	}

	handler := func(event *api.EventDTO) error {

		endDate := utils.MillisToTimeUTC(event.EndDate)

		if endDate.After(currentTime) {
			log.Printf("WARNING: Stop archiving events because there are events that have not finished yet (%v)", event.Id)
			return ErrCannotArchive
		}

		if event.Cancelled {
			// Skip cancelled events. They are archived when cancelled.
			log.Printf("WARNING: Skip archiving event %v because it's cancelled (so likely already archived)", event.Id)
			return nil
		}

		// Archive
		if err := m.archiveEvent(event); err != nil {
			return err
		}

		// Remove from inbox
		for pID := range event.Participants {
			m.userEvents.Remove(pID, event.Id)
		}

		return nil
	}

	if len(entries) > 0 {

		log.Printf("Num.Events %v from %v to %v", len(entries), from, currentTime)

		// Extract IDs
		eventIDs := make([]int64, 0, len(entries))
		for _, entry := range entries {
			eventIDs = append(eventIDs, entry.EventID)
		}

		// Range events
		// TODO: If archive fails in the middle then retry from (minute, event). Currently, retry policy starts
		// again from (minute) so it will re-archive already archived events. This is not harmful but is a
		// waste of resources.
		err = m.eventDAO.RangeEvents(handler, eventIDs...)
		if err != nil {
			return err
		}

		// Update lastArchiveTime
		value := strconv.FormatInt(utils.TimeToMillis(currentTime), 10)
		if err := m.settingsDAO.Insert(api.MasterLastArchiveTime, value); err != nil {
			return err
		}
		m.lastArchiveTime = currentTime
	}

	return nil
}
