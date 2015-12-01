package main

import (
	"github.com/twinj/uuid"
	"log"
)

func NewUserDatabase() *UsersDatabase {
	udb := &UsersDatabase{}
	udb.allusers = make(map[string]*UserAccount)
	udb.facebook = make(map[string]*UserAccount)
	udb.idIndex = make(map[uint64]*UserAccount)
	return udb
}

type UsersDatabase struct {
	allusers map[string]*UserAccount // Index by E-mail
	facebook map[string]*UserAccount // Index by Facebok User ID
	idIndex  map[uint64]*UserAccount // Index by UserID
}

// Checks if an email exists
func (udb *UsersDatabase) ExistEmail(email string) bool {
	_, ok := udb.allusers[email]
	return ok
}

// Checks if an fbid exists
func (udb *UsersDatabase) ExistFB(fbid string) bool {
	_, ok := udb.facebook[fbid]
	return ok
}

// Checks if an User ID exists
func (udb *UsersDatabase) ExistID(id uint64) bool {
	_, ok := udb.idIndex[id]
	return ok
}

// Check if a slice of User ID exists
func (udb *UsersDatabase) ExistAllIDs(ids []uint64) bool {
	// Check valid user participants
	is_valid := true

	for _, user_id := range ids {
		if !udb.ExistID(user_id) {
			is_valid = false
			break
		}
	}

	return is_valid
}

// Checks if a given user id with auth_token has access
func (udb *UsersDatabase) CheckAccess(id uint64, auth_token uuid.UUID) bool {

	result := false

	if auth_token != nil {
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
func (udb *UsersDatabase) GetByID(id uint64) (uac *UserAccount, ok bool) {
	uac, ok = udb.idIndex[id]
	return
}

// Get an user account by FB User ID
func (udb *UsersDatabase) GetByFBUID(fbid string) (uac *UserAccount, ok bool) {
	uac, ok = udb.facebook[fbid]
	return
}

// Insert a new user into the database
func (udb *UsersDatabase) Insert(account *UserAccount) bool {

	if udb.ExistEmail(account.email) {
		log.Println("Given account (", account.email, ") already exists")
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
	udb.idIndex[account.id] = account
	account.udb = udb

	return true
}

// Removes an user from the database
func (udb *UsersDatabase) Remove(email string) {

	if udb.ExistEmail(email) {
		account := udb.allusers[email]
		delete(udb.allusers, email)
		delete(udb.idIndex, account.id)
		if account.IsFacebook() {
			delete(udb.facebook, account.fbid)
		}
	}
}
