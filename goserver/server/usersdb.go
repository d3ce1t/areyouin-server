package main

import (
	"github.com/twinj/uuid"
	"log"
)

type UsersDatabase struct {
	allusers map[string]*UserAccount // Index by E-mail
	facebook map[string]*UserAccount // Index by Facebok User ID
	idIndex  map[string]*UserAccount // Index by UserID
}

// Checks if an email exists
func (udb *UsersDatabase) Exist(email string) bool {
	_, ok := udb.allusers[email]
	return ok
}

// Checks if an fbod exists
func (udb *UsersDatabase) ExistFB(fbid string) bool {
	_, ok := udb.facebook[fbid]
	return ok
}

// Checks if a given user id with auth_token has access
func (udb *UsersDatabase) CheckAccess(id uuid.UUID, auth_token uuid.UUID) bool {

	result := false

	if id != nil && auth_token != nil {
		if userAccount, ok := udb.GetByID(id); ok {
			if uuid.Equal(auth_token, userAccount.auth_token) {
				result = true
			}
		}
	}

	return result
}

// Get an user account by e-mail
func (udb *UsersDatabase) GetByEmail(email string) (uac *UserAccount, ok bool) {
	uac, ok = udb.allusers[email]
	return
}

// Get an user account by User ID
func (udb *UsersDatabase) GetByID(id uuid.UUID) (uac *UserAccount, ok bool) {
	uac, ok = udb.idIndex[id.String()]
	return
}

// Get an user account by FB User ID
func (udb *UsersDatabase) GetByFBUID(fbid string) (uac *UserAccount, ok bool) {
	uac, ok = udb.facebook[fbid]
	return
}

// Insert a new user into the database
func (udb *UsersDatabase) Insert(account *UserAccount) bool {

	if udb.Exist(account.email) {
		return false
	}

	if account.IsFacebook() {
		if _, exist := udb.facebook[account.fbid]; exist {
			log.Println("FIXME: Facebook ID already exists. This isn't supposed to happen but it has")
			return false
			// FIXME: Can I recover? May be overwrite anyway?
		}

		udb.facebook[account.fbid] = account
	}

	udb.allusers[account.email] = account
	udb.idIndex[account.id.String()] = account

	return true
}

// Removes an user from the database
func (udb *UsersDatabase) Remove(email string) {

	if udb.Exist(email) {
		account := udb.allusers[email]
		delete(udb.allusers, email)
		delete(udb.idIndex, account.id.String())
		if account.IsFacebook() {
			delete(udb.facebook, account.fbid)
		}
	}
}

func newUserDatabase() *UsersDatabase {
	udb := &UsersDatabase{}
	udb.allusers = make(map[string]*UserAccount)
	udb.facebook = make(map[string]*UserAccount)
	udb.idIndex = make(map[string]*UserAccount)
	return udb
}
