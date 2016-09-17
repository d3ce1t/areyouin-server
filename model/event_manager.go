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
	dbsession   api.DbSession
	parent      *AyiModel
	userDAO     api.UserDAO
	eventDAO    api.EventDAO
	thumbDAO    api.ThumbnailDAO
	eventSignal observer.Property
}

func newEventManager(parent *AyiModel, session api.DbSession) *EventManager {
	return &EventManager{
		parent:      parent,
		dbsession:   session,
		userDAO:     cqldao.NewUserDAO(session),
		eventDAO:    cqldao.NewEventDAO(session),
		thumbDAO:    cqldao.NewThumbnailDAO(session),
		eventSignal: observer.NewProperty(nil),
	}
}

func (m *EventManager) Observe() observer.Stream {
	return m.eventSignal.Observe()
}

// CreateNewEvent creates an event with no participants
//
// Prominent errors:
// - ErrInvalidAuthor
// - ErrEventOutOfCreationWindow
// - ErrInvalidEventData
// - ErrInvalidStartDate
// - ErrInvalidEndDate
//
// Assumes:
// (1) author must exists and be valid
//
// Preconditions:
// (1) event created date must be inside a valid window
// (2) event start and end date must obey business rules
func (m *EventManager) CreateNewEvent(author *UserAccount, createdDate int64,
	startDate int64, endDate int64, description string) (*Event, error) {

	// Create event
	event := NewEvent(author.id, author.name, createdDate, startDate, endDate, description)

	// Check precondition (3)

	if _, err := event.IsValid(); err != nil {
		return nil, err
	}

	// Check precondition (2)
	currentDateTime := utils.UnixMillisToTimeUTC(utils.GetCurrentTimeSeconds())
	createdDateTime := utils.UnixMillisToTimeUTC(event.CreatedDate())

	if createdDateTime.Before(currentDateTime.Add(-time.Minute)) || createdDateTime.After(currentDateTime.Add(time.Minute)) {
		return nil, ErrEventOutOfCreationWindow
	}

	return event, nil
}

// CreateParticipantsList creates a participant list by means of the provided participants id.
// Only friends of user 'authorId' are included in the resulting list.
func (m *EventManager) CreateParticipantsList(authorID int64, participants []int64) (map[int64]*UserAccount, error) {

	if len(participants) == 0 {
		return nil, ErrParticipantsRequired
	}

	usersList := make(map[int64]*UserAccount)
	friendsCounter := 0

	for _, pID := range participants {

		if ok, err := m.parent.Friends.IsFriend(pID, authorID); ok {

			friendsCounter++

			// Participant has author as his friend

			user, err := m.userDAO.Load(pID)
			if err == api.ErrNotFound {

				// Participant doesn't exist

				// TODO: Send e-mail to Admin
				log.Printf("* CREATE PARTICIPANT LIST WARNING: USER %v NOT FOUND: This means user (%v) has a friend list but its account doesn't exist. Admin required.\n", pID, pID)
				continue

			} else if err != nil {
				return nil, err
			}

			usersList[user.Id] = newUserFromDTO(user)

		} else if err != nil {
			return nil, err
		} else {
			log.Printf("* CREATE PARTICIPANT LIST WARNING: USER %v TRIED TO ADD USER %v BUT THEY ARE NOT FRIENDS\n", authorID, pID)
		}

	} // End for

	if len(usersList) == 0 {
		if friendsCounter > 0 {
			return nil, ErrModelInconsistency
		} else {
			return nil, ErrParticipantsRequired
		}
	}

	return usersList, nil
}

// Publish an event, i.e. store it in such a way that if a participant request
// his events list, the event will be included.
//
// Note that event should not have participants stored in his event.Participants
// field. This method will set this property to include only those users whose
// event delivery had succeeded.
//
// Preconditions:
// (1) author must exists and be valid
// (2) event created date must be inside a valid window
// (3) event start and end date must obey business rules
// (4) Should be at least 1 participant besides the author
func (m *EventManager) PublishEvent(event *Event, users map[int64]*UserAccount) error {

	// Check precondition (3)

	if _, err := event.IsValid(); err != nil {
		return err
	}

	// Check precondition (4)

	if len(users) == 0 {
		return ErrParticipantsRequired
	}

	// Check precondition (2)

	currentDate := utils.UnixMillisToTimeUTC(utils.GetCurrentTimeSeconds())
	createdDate := utils.UnixMillisToTimeUTC(event.CreatedDate())

	if createdDate.Before(currentDate.Add(-time.Minute)) || createdDate.After(currentDate.Add(time.Minute)) {
		return ErrEventOutOfCreationWindow
	}

	// Check precondition (1)

	authorDto, err := m.userDAO.Load(event.AuthorId())
	if err == api.ErrNotFound {
		return ErrInvalidAuthor
	} else if err != nil {
		return err
	}

	author := newUserFromDTO(authorDto)

	// Store event: If suceed, it's guaranteed that event exists, author is one of
	// the participants, and author has the event in his events inbox.
	// Otherwise, if it fails, an event with no participants may exist (orphaned event)
	err = m.persistEvent(event, author.AsParticipant())
	if err != nil {
		return err
	}

	// Emit signal
	signal := &Signal{
		Type: SignalNewEvent,
		Data: map[string]interface{}{
			"EventID": event.Id(),
			"Event":   event,
		},
	}

	m.eventSignal.Update(signal)

	// Invite users. If error do nothing.
	m.InviteUsers(event, users)

	return nil
}

// Cancel an existing event_ids
//
// Assumptions:
// - (1) Event exist and is persisted in DB
// - (2) User who performs this operation have permission
//
// Preconditions:
// - (1) Event must have not started
func (m *EventManager) CancelEvent(event *Event, userID int64) error {

	// Check precondition (1)
	if event.Status() != api.EventState_NOT_STARTED {
		return ErrEventNotWritable
	}

	// Change event state and position in time line
	new_inbox_position := utils.GetCurrentTimeMillis()
	err := m.eventDAO.SetEventStateAndInboxPosition(event.id, api.EventState_CANCELLED,
		new_inbox_position)

	if err != nil {
		return err
	}

	// Moreover, change event position inside user inbox so that next time client request recent events
	// this one is ignored.
	/*for _, participant := range event.Participants {
		err := eventDAO.SetUserEventInboxPosition(participant, event, new_inbox_position)
		if err != nil {
			log.Println("onCancelEvent Error:", err) // FIXME: Add retry logic
		}
	}*/

	// Update event object
	event.inboxPosition = new_inbox_position
	event.cancelled = true

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

// Invite participants to an existing event.
// Returns a slice of participants that were actually invited.
//
// TODO: By now, only author can invite participants to an event.
// However, when other users are able of doing the same it will be
// needed to consider concurrency issues that may cause data
// inconsistency.
//
// Assumptions:
// - (1) Event exist and is persisted in DB
// - (2) User who performs this operation have permission
//
// Preconditions:
// - (1) Event must have not started
// - (2) There must to be at least one new participant to be invited
func (m *EventManager) InviteUsers(event *Event, users map[int64]*UserAccount) (map[int64]*UserAccount, error) {

	// Check precondition (1)
	if event.Status() != api.EventState_NOT_STARTED {
		return nil, ErrEventNotWritable
	}

	// Check precondition (2) and invite participants at the same time
	oldParticipants := event.ParticipantIds()
	newParticipants := 0
	usersInvited := make(map[int64]*UserAccount)

	for _, user := range users {

		if _, ok := event.participants[user.id]; !ok {

			newParticipants++

			participant := user.AsParticipant()
			participant.invitationStatus = api.InvitationStatus_SERVER_DELIVERED

			if err := m.eventDAO.AddParticipantToEvent(participant.AsDTO(), event.AsDTO()); err == nil {
				event.addParticipant(participant)
				usersInvited[user.id] = user
			} else {
				log.Printf("* INVITE USERS WARNING: %v\n", err)
			}
		}
	}

	if newParticipants == 0 || len(usersInvited) == 0 {
		log.Printf("* INVITE USERS WARNING: %v/%v participants added\n", len(usersInvited), newParticipants)
		return nil, ErrParticipantsRequired
	}

	// TODO: If two users invite participants at the same time, counter will be inconsistent.
	// Remove this implementation when possible. Replace by another one where counters aren't needed
	if _, err := m.eventDAO.SetNumGuests(event.id, len(event.participants)); err != nil {
		log.Printf("* INVITE USERS WARNING: Update Num. guestss Error %v\n", err)
	}

	event.numGuests = int32(len(event.participants))

	// Emit signal
	signal := &Signal{
		Type: SignalEventParticipantsInvited,
		Data: map[string]interface{}{
			"EventID":         event.Id(),
			"NewParticipants": GetUserMapKeys(usersInvited),
			"OldParticipants": oldParticipants,
			"Event":           event,
		},
	}

	m.eventSignal.Update(signal)

	return usersInvited, nil
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

	if picture != nil && len(picture) != 0 {

		// Set event picture

		// Compute digest for picture
		digest := sha256.Sum256(picture)

		corePicture := &Picture{
			RawData: picture,
			Digest:  digest[:],
		}

		// Save event picture
		if err := m.saveEventPicture(event.id, corePicture); err != nil {
			return err
		}

		event.pictureDigest = corePicture.Digest

	} else {

		// Remove event picture

		if err := m.removeEventPicture(event.id); err != nil {
			return err
		}

		event.pictureDigest = nil
	}

	// Emit signal
	signal := &Signal{
		Type: SignalEventInfoChanged,
		Data: map[string]interface{}{
			"EventID": event.Id(),
			"Event":   event,
		},
	}

	m.eventSignal.Update(signal)

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

func (m *EventManager) GetEvent(eventID int64) (*Event, error) {

	events, err := m.eventDAO.LoadEvents(eventID)
	if err != nil {
		return nil, err
	}

	if len(events) > 0 {
		return newEventFromDTO(events[0]), nil
	}

	return nil, ErrNotFound
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
func (m *EventManager) GetRecentEvents(userId int64) ([]*Event, error) {

	currentTime := utils.GetCurrentTimeMillis()
	eventsDto, err := m.eventDAO.LoadRecentEventsFromUser(userId, currentTime)

	if err == api.ErrNoResults {
		return nil, ErrEmptyInbox
	} else if err != nil {
		return nil, err
	}

	events := make([]*Event, 0, len(eventsDto))
	for _, ev := range eventsDto {
		events = append(events, newEventFromDTO(ev))
	}

	return events, nil
}

func (m *EventManager) GetEventsHistory(userId int64, start int64, end int64) ([]*Event, error) {

	currentTime := utils.GetCurrentTimeMillis()

	if start >= currentTime {
		start = currentTime
	}

	if end >= currentTime {
		end = currentTime
	}

	eventsDto, err := m.eventDAO.LoadEventsHistoryFromUser(userId, start, end)
	if err == api.ErrNoResults {
		return nil, ErrEmptyInbox
	} else if err != nil {
		return nil, err
	}

	events := make([]*Event, 0, len(eventsDto))
	for _, ev := range eventsDto {
		events = append(events, newEventFromDTO(ev))
	}

	return events, nil
}

/*func (self *EventManager) FilterEvents(events []*Event, targetParticipant int64) []*Event {

	filteredEvents := make([]*Event, 0, len(events))

	for _, event := range events {
		filteredEvents = append(filteredEvents, self.GetFilteredEvent(event, targetParticipant))
	}

	return filteredEvents
}*/

/*func (self *EventManager) filterParticipants(participants map[int64]*Participant, targetParticipant int64) map[int64]*Participant {

	result := make(map[int64]*Participant)

	for key, p := range participants {
		if ok, _ := self.canSee(targetParticipant, p); ok {
			result[key] = p
		} else {
			result[key] = p.asAnonym()
		}
	}

	return result
}*/

/*func (self *EventManager) FilterParticipantsSlice(participants []*Participant, targetParticipant int64) map[int64]*Participant {

	result := make(map[int64]*Participant)

	for _, p := range participants {
		if ok, _ := self.canSee(targetParticipant, p); ok {
			result[p.id] = p
		} else {
			result[p.id] = p.asAnonym()
		}
	}

	return result
}*/

/*func (self *EventManager) GetFilteredEvent(event *Event, targetParticipant int64) *Event {

	// Clone event with empty participant list (num.attendees and num.guest are preserved)
	eventCopy := event.CloneEmptyParticipants()

	// Copy filtered participants list to the new event
	eventCopy.participants = self.GetFilteredParticipants(event, targetParticipant)
	return eventCopy
}*/

/*func (self *EventManager) GetFilteredParticipants(event *Event, targetParticipant int64) map[int64]*Participant {

	if len(event.participants) == 0 {
		log.Printf("FILTER PARTICIPANTS WARNING: Event %v has zero participants\n", event.id)
		return nil
	}

	return self.filterParticipants(event.participants, targetParticipant)
}*/

// Insert an event into database, add participants to it and send it to users' inbox.
func (m *EventManager) persistEvent(event *Event, author *Participant) error {

	if event.AuthorId() != author.Id() {
		return ErrInvalidAuthor
	}

	if err := m.eventDAO.Insert(event.AsDTO()); err != nil {
		return err
	}

	// Add author first in order to let author receive the event and add other
	// participants if something fails
	author.response = api.AttendanceResponse_ASSIST

	if err := m.eventDAO.AddParticipantToEvent(author.AsDTO(), event.AsDTO()); err != nil {
		return ErrAuthorDeliveryError
	}

	// TODO: NumAttendees and NumGuests isn't updated

	event.addParticipant(author)

	return nil
}

func (m *EventManager) saveEventPicture(event_id int64, picture *Picture) error {

	// Decode image
	srcImage, _, err := image.Decode(bytes.NewReader(picture.RawData))
	if err != nil {
		return err
	}

	// Check image size is inside bounds
	if srcImage.Bounds().Dx() > EVENT_PICTURE_MAX_WIDTH || srcImage.Bounds().Dy() > EVENT_PICTURE_MAX_HEIGHT {
		return ErrImageOutOfBounds
	}

	// Create thumbnails
	thumbnails, err := utils.CreateThumbnails(srcImage, EVENT_THUMBNAIL, m.parent.supportedDpi)
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
