package model

import (
  "bytes"
  core "peeple/areyouin/common"
  "peeple/areyouin/dao"
  "crypto/sha256"
  "image"
  "time"
  "log"
)

func newEventManager(parent *AyiModel, session core.DbSession) *EventManager {
  return &EventManager{
    parent: parent,
    dbsession: session,
    userDAO: dao.NewUserDAO(session),
    eventDAO: dao.NewEventDAO(session),
  }
}

type EventManager struct {
  dbsession core.DbSession
  parent *AyiModel
  userDAO core.UserDAO
  eventDAO core.EventDAO
}

// Prominent errors:
// - ErrInvalidAuthor
// - ErrEventOutOfCreationWindow
// - ErrInvalidEventData
// - ErrInvalidStartDate
// - ErrInvalidEndDate
//
// Preconditions:
// (1) author must exists and be valid
// (2) event created date must be inside a valid window
// (3) event start and end date must obey business rules
func (self *EventManager) CreateNewEvent(author *core.UserAccount, createdDate int64, startDate int64, endDate int64, description string) (*core.Event, error) {

  // Create event
  event := core.CreateNewEvent(author.Id, author.Name, createdDate, startDate, endDate, description)

  // Check precondition (3)

  if _, err := event.IsValid(); err != nil {
    return nil, err
  }

  // Check precondition (1)

  valid, err := self.userDAO.CheckValidAccountObject(author.Id, author.Email, author.Fbid, true)
  if err != nil {
    return nil, err
  }

  if !valid {
    return nil, ErrInvalidAuthor
  }

  // Check precondition (2)
  currentDateTime := core.UnixMillisToTime(core.GetCurrentTimeSeconds())
	createdDateTime := core.UnixMillisToTime(event.CreatedDate)

  if createdDateTime.Before(currentDateTime.Add(-time.Minute)) || createdDateTime.After(currentDateTime.Add(time.Minute)) {
		return nil, ErrEventOutOfCreationWindow
	}

  return event, nil
}

// Create a participant list by means of the provided participants id.
// Only friends of user with authorId are included in the resulting list.
func (self *EventManager) CreateParticipantsList(authorId int64, participants []int64) (map[int64]*core.UserAccount, error) {

  if len(participants) == 0 {
		return nil, ErrParticipantsRequired
	}

  usersList := make(map[int64]*core.UserAccount)
  friendsCounter := 0

  for _, p_id := range participants {

    if ok, err := self.parent.Accounts.IsFriend(p_id, authorId); ok {

      friendsCounter++

      // Participant has author as his friend

      user, err := self.userDAO.Load(p_id)
      if err == dao.ErrNotFound {

        // Participant doesn't exist

        // TODO: Send e-mail to Admin
        log.Printf("* CREATE PARTICIPANT LIST WARNING: USER %v NOT FOUND: This means user (%v) has friend list but it doesn't exist. Admin required.\n", p_id, p_id)
        continue

      } else if err != nil {
        return nil, err
      }

      usersList[user.Id] = user

    } else if err != nil {
      return nil, err
    } else {
      log.Printf("* CREATE PARTICIPANT LIST WARNING: USER %v TRIED TO ADD USER %v BUT THEY ARE NOT FRIENDS\n", authorId, p_id)
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
func (self *EventManager) PublishEvent(event *core.Event, users map[int64]*core.UserAccount) error {

  // Check precondition (3)

  if _, err := event.IsValid(); err != nil {
    return err
  }

  // Check precondition (4)

  if len(users) == 0 {
		return ErrParticipantsRequired
	}

  // Check precondition (2)

  currentDate := core.UnixMillisToTime(core.GetCurrentTimeSeconds())
	createdDate := core.UnixMillisToTime(event.CreatedDate)

  if createdDate.Before(currentDate.Add(-time.Minute)) || createdDate.After(currentDate.Add(time.Minute)) {
		return ErrEventOutOfCreationWindow
	}

  // Check precondition (1)

  author, err := self.userDAO.Load(event.AuthorId)
  if err == dao.ErrNotFound {
    return ErrInvalidAuthor
  }

  valid, err := self.userDAO.CheckValidAccountObject(author.Id, author.Email, author.Fbid, true)
  if err != nil {
    return err
  }

  if !valid {
    return ErrInvalidAuthor
  }

  // Store event: If suceed, it's guaranteed that event exists, author is one of
  // the participants, and author has the event in his events inbox.
  // Otherwise, if it fails, an event with no participants may exist (orphaned event)
  err = self.persistEvent(event, author.AsParticipant())
  if err != nil {
    return err
  }

  // Invite users. If error do nothing.
  self.InviteUsers(event, users)
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
func (self *EventManager) CancelEvent(event *core.Event) error {

  // Check precondition (1)
  if event.GetStatus() != core.EventState_NOT_STARTED {
    return ErrEventNotWritable
  }

  // Change event state and position in time line
	new_inbox_position := core.GetCurrentTimeMillis()
	err := self.eventDAO.SetEventStateAndInboxPosition(
    event.EventId, core.EventState_CANCELLED,
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
	event.InboxPosition = new_inbox_position
	event.State = core.EventState_CANCELLED

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
func (self *EventManager) InviteUsers(event *core.Event, users map[int64]*core.UserAccount) (map[int64]*core.UserAccount, error) {

  // Check precondition (1)
  if event.GetStatus() != core.EventState_NOT_STARTED {
    return nil, ErrEventNotWritable
  }

  // Check precondition (2) and invite participants at the same time
  newParticipants := 0
  addedParticipants := 0
  usersInvited := make(map[int64]*core.UserAccount)


  for _, user := range users {
    if _, ok := event.Participants[user.Id]; !ok {

      newParticipants++

      // TODO: Implement retry but do not return on error
      participant := user.AsParticipant()
      participant.Delivered = core.MessageStatus_SERVER_DELIVERED

      if err := self.eventDAO.AddParticipantToEvent(participant, event); err == nil {
        event.AddParticipant(participant)
        usersInvited[user.Id] = user
        addedParticipants++
      } else {
        log.Printf("* INVITE USERS WARNING: %v\n", err)
      }
    }
  }

  if newParticipants == 0 || addedParticipants == 0 {
    log.Printf("* INVITE USERS WARNING: %v/%v participants added\n", addedParticipants, newParticipants)
    return nil, ErrParticipantsRequired
  }

  if _, err := self.eventDAO.CompareAndSetNumGuests(event.EventId, len(event.Participants)); err != nil {
    log.Printf("* INVITE USERS WARNING: Update Num. guestss Error %v\n", err)
  }

  event.NumGuests = int32(len(event.Participants))

  return usersInvited, nil
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
func (self *EventManager) ChangeParticipantResponse(userId int64, response core.AttendanceResponse, event *core.Event) (bool, error) {

  // Check precondition (1)

  if event.GetStatus() != core.EventState_NOT_STARTED {
    return false, ErrEventNotWritable
  }

  // Check precondition (2)

  participant, ok := event.Participants[userId]
  if !ok {
    return false, ErrParticipantNotFound
  }

	if participant.Response != response {

    // Change response

    if err := self.eventDAO.SetParticipantResponse(participant.UserId, response, event); err != nil {
  	   return false, err
  	}

    participant.Response = response
    return true, nil
	}

  return false, nil
}

// Change event picture
//
// Assumptions:
// - (1) Event exist and is persisted in DB
// - (2) User who performs this operation have permission
//
// Preconditions
// - (1) Event must have not started
func (self *EventManager) ChangeEventPicture(event *core.Event, picture []byte) error {

  // Check precondition (1)
  if event.GetStatus() != core.EventState_NOT_STARTED {
    return ErrEventNotWritable
  }

  if picture != nil && len(picture) != 0 {

    // Set event picture

    // Compute digest for picture
    digest := sha256.Sum256(picture)

    corePicture := &core.Picture{
      RawData: picture,
      Digest:  digest[:],
    }

    // Save event picture
    if err := self.saveEventPicture(event.EventId, corePicture); err != nil {
      return err
    }

    event.PictureDigest = corePicture.Digest

  } else {

    // Remove event picture

    if err := self.removeEventPicture(event.EventId); err != nil {
      return err
    }

    event.PictureDigest = nil
  }

  return nil
}

func (self *EventManager) ChangeDeliveryState(event *core.Event, userId int64, state core.MessageStatus) (bool, error) {

  if event.GetStatus() != core.EventState_NOT_STARTED {
    return false, ErrEventNotWritable
  }

  participant, ok := event.Participants[userId]

  if ok {

    if participant.Delivered == state {
      return false, nil // Do nothing
    }

    // TODO: Add business rule to avoid move to a previous state

    err := self.eventDAO.SetParticipantStatus(userId,
      event.EventId, state)

    if err != nil {
      return false, err
    }

    participant.Delivered = state
    return true, nil

  } else {
    return false, ErrParticipantNotFound
  }
}

// FIXME: Do not get all of the private events, but limit to
// a fixed number.
func (self *EventManager) GetRecentEvents(userId int64) ([]*core.Event, error) {

  current_time := core.GetCurrentTimeMillis()
	events, err := self.eventDAO.LoadUserEventsAndParticipants(userId, current_time)
  if err == dao.ErrNoResults {
    return nil, ErrEmptyInbox
	} else if err != nil {
    return nil, err
  }

  return events, nil
}

func (self *EventManager) GetEventsHistory(userId int64, start int64, end int64) ([]*core.Event, error) {

  current_time := core.GetCurrentTimeMillis()

	if start >= current_time {
		start = current_time
	}

	if end >= current_time {
		end = current_time
	}

  return self.eventDAO.LoadUserEventsHistoryAndparticipants(userId, start, end)
}

func (self *EventManager) FilterEvents(events []*core.Event, targetParticipant int64) []*core.Event {

  filteredEvents := make([]*core.Event, 0, len(events))

  for _, event := range events {
    filteredEvents = append(filteredEvents, self.GetFilteredEvent(event, targetParticipant))
  }

  return filteredEvents
}

func (self *EventManager) FilterParticipants(participants map[int64]*core.EventParticipant, targetParticipant int64) map[int64]*core.EventParticipant {

  result := make(map[int64]*core.EventParticipant)

	for key, p := range participants {
		if ok, _ := self.canSee(targetParticipant, p); ok {
			result[key] = p
		} else {
			result[key] = p.AsAnonym()
		}
	}

  return result
}

func (self *EventManager) FilterParticipantsSlice(participants []*core.EventParticipant, targetParticipant int64) map[int64]*core.EventParticipant {

  result := make(map[int64]*core.EventParticipant)

	for _, p := range participants {
		if ok, _ := self.canSee(targetParticipant, p); ok {
			result[p.UserId] = p
		} else {
			result[p.UserId] = p.AsAnonym()
		}
	}

  return result
}

func (self *EventManager) GetFilteredEvent(event *core.Event, targetParticipant int64) *core.Event {

  // Clone event with empty participant list (num.attendees and num.guest are preserved)
  eventCopy := event.CloneEmpty()

  // Copy filtered participants list to the new event
  participants := self.GetFilteredParticipants(event, targetParticipant)
  eventCopy.SetParticipants(participants)

	return eventCopy
}

func (self *EventManager) GetFilteredParticipants(event *core.Event, targetParticipant int64) map[int64]*core.EventParticipant {

  if len(event.Participants) == 0 {
    log.Printf("FILTER PARTICIPANTS WARNING: Event %v has zero participants\n", event.EventId)
    return nil
  }

  return self.FilterParticipants(event.Participants, targetParticipant)
}

// Insert an event into database, add participants to it and send it to users' inbox.
func (self *EventManager) persistEvent(event *core.Event, author *core.EventParticipant) error {

  if event.AuthorId != author.UserId {
    return ErrInvalidAuthor
  }

	if err := self.eventDAO.InsertEvent(event); err != nil {
		return err
	}

	// Add author first in order to let author receive the event and add other
  // participants if something fails
  author.Response = core.AttendanceResponse_ASSIST
  authorCopy := *author
  authorCopy.Delivered = core.MessageStatus_CLIENT_DELIVERED

  if err := self.eventDAO.AddParticipantToEvent(&authorCopy, event); err != nil {
    return ErrAuthorDeliveryError
  }

  event.AddParticipant(author)

	return nil
}

func (self *EventManager) saveEventPicture(event_id int64, picture *core.Picture) error {

	// Decode image
	srcImage, _, err := image.Decode(bytes.NewReader(picture.RawData))
	if err != nil {
		return err
	}

	// Check image size is inside bounds
	if srcImage.Bounds().Dx() > core.EVENT_PICTURE_MAX_WIDTH || srcImage.Bounds().Dy() > core.EVENT_PICTURE_MAX_HEIGHT {
		return ErrImageOutOfBounds
	}

	// Create thumbnails
	thumbnails, err := core.CreateThumbnails(srcImage, EVENT_THUMBNAIL, self.parent.supportedDpi)
	if err != nil {
		return err
	}

	// Save thumbnails
	thumbDAO := dao.NewThumbnailDAO(self.dbsession)
	err = thumbDAO.Insert(event_id, picture.Digest, thumbnails)
	if err != nil {
		return err
	}

	// Save event picture (always does it after thumbnails)
	eventDAO := dao.NewEventDAO(self.dbsession)
	err = eventDAO.SetEventPicture(event_id, picture)
	if err != nil {
		return err
	}

	return nil
}

func (self *EventManager) removeEventPicture(event_id int64) error {

	// Remove event picture
	eventDAO := dao.NewEventDAO(self.dbsession)
	err := eventDAO.SetEventPicture(event_id, &core.Picture{RawData: nil, Digest:  nil})
	if err != nil {
		return err
	}

	// Remove thumbnails
	thumbDAO := dao.NewThumbnailDAO(self.dbsession)
	err = thumbDAO.Remove(event_id)
	if err != nil {
		return err
	}

	return nil
}

// Tells if participant p1 can see event changes of participant p2
func (self *EventManager) canSee(p1 int64, p2 *core.EventParticipant) (bool, error) {
	if p2.Response == core.AttendanceResponse_ASSIST {
		return true, nil
  } else {
    return self.parent.Accounts.IsFriend(p2.UserId, p1)
	}
}
