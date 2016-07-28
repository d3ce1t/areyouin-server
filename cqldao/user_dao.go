package cqldao

import (
	"log"
	"peeple/areyouin/api"

	"github.com/gocql/gocql"
)

const (
	GRACE_PERIOD_MS = 20 * 1000 // 20s
)

type UserDAO struct {
	session *GocqlSession
}

// TODO: Include Password and Salt in user_account table
// TODO: Make up some type of high level iterator in order to control scanning
// steps in model layer.
func (d *UserDAO) LoadAll() ([]*api.UserDTO, error) {

	checkSession(d.session)

	stmt := `SELECT user_id, auth_token, email, email_verified, name, fb_id, fb_token,
						iid_token, network_version, last_connection, created_date, picture_digest
						FROM user_account LIMIT 2000`

	iter := d.session.Query(stmt).Iter()

	if iter == nil {
		return nil, ErrUnexpected
	}

	users := make([]*api.UserDTO, 0, 1024)
	var dto api.UserDTO

	for iter.Scan(&dto.Id, &dto.AuthToken, &dto.Email, &dto.EmailVerified, &dto.Name,
		&dto.FbId, &dto.FbToken, &dto.IidToken.Token, &dto.IidToken.Version, &dto.LastConn,
		&dto.CreatedDate, &dto.PictureDigest) {
		userDTO := new(api.UserDTO)
		*userDTO = dto
		users = append(users, userDTO)
	}

	if err := iter.Close(); err != nil {
		return nil, convErr(err)
	}

	return users, nil
}

func (d *UserDAO) Load(userId int64) (*api.UserDTO, error) {

	// Load account
	user, err := d.loadUserAccount(userId)
	if err != nil {
		return nil, err
	}

	// Check that has an email. If it doesn't exist, do not consider it as an inconsistency.
	// That account is in an invalid state (but managed), i.e. account exist, probably fb,
	// but not e-mail
	cred, err := d.loadEmailCredential(user.Email)
	if err == api.ErrNotFound {
		// TODO: Send e-mail to Admin
		log.Printf("* LOAD WARNING: EMAIL %v NOT FOUND: This means an user (%v) exists but has Email (%v) that does not exist\n", user.Email, user.Id, user.Email)
		return nil, ErrInconsistency
	} else if err != nil {
		return nil, err
	}

	// Most importantly, check that e_mail credential belongs to given user_id
	if user.Id != cred.UserId {
		// TODO: Send e-mail to Admin
		log.Printf("* LOAD WARNING: USER %v MISMATCH: This means an AyiUser (%v) points to an Email (%v) that does point to another AyiUser (%v)\n", user.Id, user.Id, user.Email, cred.UserId)
		return nil, ErrInconsistency
	}

	user.Password = cred.Password
	user.Salt = cred.Salt

	return user, nil
}

func (d *UserDAO) LoadByEmail(email string) (*api.UserDTO, error) {

	checkSession(d.session)

	if email == "" {
		return nil, api.ErrNotFound
	}

	cred, err := d.loadEmailCredential(email)
	if err != nil {
		return nil, err
	}

	user, err := d.loadUserAccount(cred.UserId)

	// Check db consistency

	if err == api.ErrNotFound {
		// TODO: Send e-mail to Admin
		log.Printf("* LOADBYEMAIL WARNING: USER %v NOT FOUND: This means an Email (%v) points to an AyiId (%v) that does not exist. Admin required.\n", cred.UserId,
			email, cred.UserId)
		return nil, ErrInconsistency
	} else if err != nil {
		return nil, err
	}

	// Check that user.Email == Email

	if user.Email != email {
		// TODO: Send e-mail to Admin
		log.Printf("* LOADBYEMAIL WARNING: USER %v EMAIL MISMATCH: This means an Email (%v) points to an AyiUser (%v) that does point to another Email (%v). Admin required.\n", user.Id, email, user.Id, user.Email)
		return nil, ErrInconsistency
	}

	user.Password = cred.Password
	user.Salt = cred.Salt

	return user, nil
}

func (d *UserDAO) LoadByFB(fbId string) (*api.UserDTO, error) {

	checkSession(d.session)

	if fbId == "" {
		return nil, api.ErrNotFound
	}

	userId, err := d.getIDByFacebookID(fbId)
	if err != nil {
		return nil, err
	}

	user, err := d.Load(userId)

	// Check db consistency. If 'not found' is returned, it is a db inconsistency.
	// That is because when creating a user account, insert order is: account,
	// fb, email. Then, if fb exists, so do account.

	if err == api.ErrNotFound {
		// TODO: Send e-mail to Admin
		log.Printf("* LOADBYFB WARNING: USER %v NOT FOUND: This means a FbId (%v) points to an AyiId (%v) that does not exist. Admin required.\n", userId, fbId, userId)
		return nil, ErrInconsistency
	} else if err == ErrInconsistency {
		// Account exist but either it has no email cred or it email cred points to another user
		return nil, api.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	//Check user.FbId == fbId

	if user.FbId != fbId {
		// TODO: Send e-mail to Admin
		log.Printf("* LOADBYFB WARNING: USER %v FBID MISMATCH: This means a FbId (%v) points to an AyiUser (%v) that does point to another FbId (%v). Admin required.\n", userId, fbId, userId, user.FbId)
		return nil, ErrInconsistency
	}

	return user, nil
}

func (d *UserDAO) LoadProfilePicture(user_id int64) (*api.PictureDTO, error) {

	checkSession(d.session)

	if user_id == 0 {
		return nil, api.ErrNotFound
	}

	stmt := `SELECT profile_picture, picture_digest
		FROM user_account	WHERE user_id = ? LIMIT 1`

	q := d.session.Query(stmt, user_id)

	picture := new(api.PictureDTO)

	err := q.Scan(&picture.RawData, &picture.Digest)
	if err != nil {
		return nil, convErr(err)
	}

	return picture, nil
}

func (d *UserDAO) LoadIIDToken(userId int64) (*api.IIDTokenDTO, error) {

	checkSession(d.session)

	if userId == 0 {
		return nil, api.ErrNotFound
	}

	stmt := `SELECT iid_token, network_version FROM user_account WHERE user_id = ?`
	q := d.session.Query(stmt, userId)

	iidToken := new(api.IIDTokenDTO)

	if err := q.Scan(&iidToken.Token, &iidToken.Version); err != nil {
		return nil, convErr(err)
	}

	return iidToken, nil
}

// Insert a new user into Cassandra involving tables user_account, user_email_credentials
// and user_facebook_credentials. It takes account race conditions like two users trying
// to create the same account simultaneously. If email is already taken this operation
// fails. In the same way, if FB account is already linked to a valid user account (row
// exists in user_account and user_email_credential), then this operation fails. With this
// implementation, orphan fb rows are removed after a grace period time in order to enable
// retry logic in the client when, by some reason, registration failed at first attempt.
// This function does not remove user account, it  only updates orphaned user_facebook_credentials
// rows that doesn't point to a valid user account.
// In order to work properly, a row in user_email_credentials must be inserted before
// GRACE_PERIOD_MS. Otherwise, another account could reclaim that FB for itself with
// another Insert call.
func (d *UserDAO) Insert(user *api.UserDTO) error {

	checkSession(d.session)

	// 1) Check if the given e-mail already exists
	if exist, err := d.existEmail(user.Email); exist {
		return api.ErrEmailAlreadyExists
	} else if err != nil {
		return err
	}

	// E-mail doesn't exist. Then continue
	insert_state := 0

	// Clean logic
	defer func() {
		switch insert_state {
		case 1:
			d.deleteUserAccount(user.Id)
		case 2:
			d.deleteFacebookCredentials(user.FbId)
			d.deleteUserAccount(user.Id) // Delete always last
		}
	}()

	// 2) Insert into user_account
	if _, err := d.insertUserAccount(user); err != nil {
		return err
	}

	// 3) Try to insert Facebook credentials considering collisions.
	// See insertFacebookCredentials for more info. If two users try to insert the same
	// FbId, only one of them will succeed.
	insert_state = 1

	if user.FbId != "" && user.FbToken != "" {
		if _, err := d.insertFacebookCredentials(user.Id, user.FbId, user.FbToken); err != nil {
			return err
		}
	}

	// 4) Finally, insert e-mail into user_email_credentials. This insert is the most
	// important because it makes valid the user account ensuring that user_account
	// and user_facebook_credentials also exist. If two users reach this point simultaneously
	// then only one of them will succeed and the other one will fail.
	insert_state = 3 // assume it's gonna succeed

	if user.Password != EMPTY_ARRAY_32B && user.Salt != EMPTY_ARRAY_32B {

		emailCred := &emailCredential{
			UserId:   user.Id,
			Email:    user.Email,
			Password: user.Password,
			Salt:     user.Salt,
		}

		if _, err := d.insertEmailCredentials(emailCred); err != nil {
			insert_state = 2
			return err
		}
	} else if _, err := d.insertEmail(user.Id, user.Email); err != nil {
		insert_state = 2
		return err
	}

	return nil
}

func (d *UserDAO) SetPassword(email string, newPassword [32]byte, newSalt [32]byte) (bool, error) {

	checkSession(d.session)

	if email == "" || newPassword == EMPTY_ARRAY_32B || newSalt == EMPTY_ARRAY_32B {
		return false, api.ErrInvalidArg
	}

	cred, err := d.loadEmailCredential(email)
	if err != nil {
		return false, err
	}

	updateEmailCredentials := `UPDATE user_email_credentials SET password = ?, salt = ?
			WHERE email = ? IF user_id = ?`

	ok, err := d.session.Query(updateEmailCredentials, newPassword[:], newSalt[:],
		cred.Email, cred.UserId).ScanCAS(nil)

	return ok, convErr(err)
}

func (d *UserDAO) SaveProfilePicture(user_id int64, picture *api.PictureDTO) error {

	checkSession(d.session)

	if user_id == 0 {
		return api.ErrInvalidArg
	}

	stmt := `UPDATE user_account SET profile_picture = ?, picture_digest = ?
						WHERE user_id = ?`

	err := d.session.Query(stmt, picture.RawData, picture.Digest, user_id).Exec()
	return convErr(err)
}

func (d *UserDAO) SetLastConnection(user_id int64, time int64) error {

	checkSession(d.session)

	if user_id == 0 {
		return api.ErrInvalidArg
	}

	stmt := `UPDATE user_account SET last_connection = ?
						WHERE user_id = ?`

	err := d.session.Query(stmt, time, user_id).Exec()
	return convErr(err)
}

func (d *UserDAO) SetAuthToken(user_id int64, auth_token string) error {

	checkSession(d.session)

	if user_id == 0 {
		return api.ErrInvalidArg
	}

	stmt := `UPDATE user_account SET auth_token = ?
						WHERE user_id = ?`
	err := d.session.Query(stmt, auth_token, user_id).Exec()
	return convErr(err)
}

func (d *UserDAO) SetFacebookToken(user_id int64, fb_id string, fb_token string) error {

	checkSession(d.session)

	if user_id == 0 {
		return api.ErrInvalidArg
	}

	batch := d.session.NewBatch(gocql.LoggedBatch) // the primary use case of a logged batch is when you need to keep tables in sync with one another, and NOT performance.

	batch.Query(`UPDATE user_facebook_credentials SET fb_token = ? WHERE fb_id = ?`,
		fb_token, fb_id)

	batch.Query(`UPDATE user_account SET fb_id = ?, fb_token = ? WHERE user_id = ?`,
		fb_id, fb_token, user_id)

	return convErr(d.session.ExecuteBatch(batch))
}

func (d *UserDAO) SetIIDToken(userID int64, iidToken *api.IIDTokenDTO) error {

	checkSession(d.session)

	if userID == 0 || iidToken == nil || iidToken.Token == "" {
		return api.ErrInvalidArg
	}

	stmt := `UPDATE user_account SET iid_token = ?, network_version = ?
						WHERE user_id = ?`
	err := d.session.Query(stmt, iidToken.Token, iidToken.Version, userID).Exec()
	return convErr(err)
}

// User information is spread in three tables: user_account, user_email_credentials
// and user_facebook_credentials. So, in order to delete a user, it's needed an
// user_id, e-mail and, likely, a Facebook ID. For the sake of safety, a read is
// perform before delete in order to perform security checks between data provided as
// argument and data stored in database. If all of the security checks passed, then user
// is removed.
func (dao *UserDAO) Delete(user *api.UserDTO) error {

	checkSession(dao.session)

	// Read

	// Security barriers
	var can_remove_email bool
	var can_remove_facebook bool

	// Read email_credential for later
	email_credential, err := dao.loadEmailCredential(user.Email)
	if err != nil {
		return err
	}

	can_remove_email = email_credential.UserId == user.Id

	if user.FbId != "" && user.FbToken != "" {
		facebook_credentials, err := dao.loadFacebookCredential(user.FbId)
		if err != nil {
			return err
		}
		can_remove_facebook = facebook_credentials.UserId == user.Id
	}

	// Read friends for deleting
	friendDAO := NewFriendDAO(dao.session)

	friends, err := friendDAO.LoadFriends(user.Id, 0)
	if err != nil {
		return err
	}

	// Read groups for deleting
	groups, err := friendDAO.LoadGroupsWithMembers(user.Id)
	if err != nil {
		return err
	}

	// Prepare Delete batch

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	// Delete groups
	for _, group := range groups {

		// Empty user's friends groups
		for _, friend_id := range group.Members {
			batch.Query(`DELETE FROM friends_by_group
				WHERE user_id = ? AND group_id = ? AND friend_id = ?`,
				user.Id, group.Id, friend_id)
		}

		// Remove friends groups
		batch.Query(`DELETE FROM groups_by_user WHERE user_id = ? AND group_id = ?`,
			user.Id, group.Id)
	}

	// Delete self user from his friends
	for _, friend := range friends {
		batch.Query(`DELETE FROM friends_by_user WHERE user_id = ? AND friend_id = ?`,
			friend.UserId, user.Id)
	}

	// Delete user's friends: Always after delete self user from other friends
	batch.Query(`DELETE FROM friends_by_user WHERE user_id = ?`, user.Id)

	// Delete email_credential. Only delete if this credential belongs to the same user
	// After this, account will be invalid
	if can_remove_email {
		batch.Query(`DELETE FROM user_email_credentials WHERE email = ?`, user.Email)
	}

	// Delete Facebook credential
	if user.FbId != "" && user.FbToken != "" && can_remove_facebook {
		batch.Query(`DELETE FROM user_facebook_credentials WHERE fb_id = ?`, user.FbId)
	}

	// Delete Thumbnails
	batch.Query(`DELETE FROM thumbnails WHERE id = ?`, user.Id)

	// Delete account. Do it always the last one operation because user_account
	// is like an index for the other information.
	batch.Query(`DELETE FROM user_account WHERE user_id = ?`, user.Id)

	return convErr(dao.session.ExecuteBatch(batch))
}
