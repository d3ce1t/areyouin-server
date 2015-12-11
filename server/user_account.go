package main

import (
	proto "areyouin/protocol"
	"github.com/twinj/uuid"
	"log"
)

func NewUserAccount(id uint64, name string, email string, password string, phone string, fbid string, fbtoken string) *UserAccount {

	user := &UserAccount{
		id:         id,
		name:       name,
		email:      email,
		password:   password,
		phone:      phone,
		fbid:       fbid,
		fbtoken:    fbtoken,
		auth_token: uuid.NewV4()}

	user.friends = make(map[uint64]*Friend)
	user.inbox = NewInbox(user.id)
	user.udb = nil

	return user
}

type UserAccount struct {
	id              uint64 // AreYouIN ID
	auth_token      uuid.UUID
	email           string
	email_verified  bool
	password        string
	name            string
	phone           string
	phone_verified  bool
	fbid            string // Facebook ID
	fbtoken         string // Facebook User Access token
	last_connection int64
	friends         map[uint64]*Friend
	udb             *UsersDatabase // Database the user belongs to
	inbox           *Inbox
}

func (ua *UserAccount) IsFacebook() bool {
	result := false
	if ua.fbid != "" && ua.fbtoken != "" {
		result = true
	}
	return result
}

func (ua *UserAccount) AddFriend(friend_id uint64) bool {

	if ua.udb == nil {
		log.Println("Trying to add a friend into a user account that has not been added to a user's database")
		return false
	}

	result := false

	if uac, ok := ua.udb.GetByID(friend_id); ok {
		ua.friends[friend_id] = &Friend{id: uac.id, name: uac.name}
		result = true
	} else {
		log.Println("Trying to add an invalid user as friend")
	}

	return result
}

func (ua *UserAccount) GetFriend(friend_id uint64) (f *Friend, ok bool) {
	f, ok = ua.friends[friend_id]
	return
}

func (ua *UserAccount) GetAllFriends() []*Friend {

	list_friends := make([]*Friend, len(ua.friends))

	i := 0
	for _, v := range ua.friends {
		list_friends[i] = v
		i++
	}

	return list_friends
}

// FIXME: It's being checked only one way, two way needed!
func (ua *UserAccount) IsFriend(friend_id uint64) bool {
	_, ok := ua.friends[friend_id]
	return ok
}

/* Returns a list of the events that the user has been invited but
   with the participants list filtered to protect privacy of non-friend
	 users*/
func (ua *UserAccount) GetAllEvents() []*proto.Event {

	list_events := make([]*proto.Event, ua.inbox.Len())

	i := 0
	for _, event := range ua.inbox.events {
		list_events[i] = &proto.Event{}
		*list_events[i] = *event // copy
		filterEventParticipants(ua, list_events[i], event)
		i++
	}

	return list_events
}
