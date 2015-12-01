package main

import (
	"github.com/twinj/uuid"
	"log"
)

func NewUserAccount(name string, email string, password string, phone string, fbid string, fbtoken string) *UserAccount {

	user := &UserAccount{
		//id:         uuid.NewV4(), DO NOT CREATE ID
		name:       name,
		email:      email,
		password:   password,
		phone:      phone,
		fbid:       fbid,
		fbtoken:    fbtoken,
		auth_token: uuid.NewV4()}

	user.friends = make(map[uint64]*Friend)
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
}

func (ua *UserAccount) IsFacebook() bool {
	result := false
	if ua.fbid != "" && ua.fbtoken != "" {
		result = true
	}
	return result
}

func (ua *UserAccount) AddFriend(friend_id uint64) bool {

	if udb == nil {
		log.Println("Trying to add a friend into a user account that has not been added to a user's database")
		return false
	}

	result := false

	if uac, ok := udb.GetByID(friend_id); ok {
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
