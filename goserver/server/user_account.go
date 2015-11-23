package main

import (
	"github.com/twinj/uuid"
)

type UserAccount struct {
	id              uuid.UUID // AreYouIN ID
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
}

func (ua *UserAccount) IsFacebook() bool {
	result := false
	if ua.fbid != "" && ua.fbtoken != "" {
		result = true
	}
	return result
}

func newUserAccount(name string, email string, password string, phone string, fbid string, fbtoken string) *UserAccount {

	user := &UserAccount{
		id:         uuid.NewV4(),
		name:       name,
		email:      email,
		password:   password,
		phone:      phone,
		fbid:       fbid,
		fbtoken:    fbtoken,
		auth_token: uuid.NewV4()}

	return user
}
