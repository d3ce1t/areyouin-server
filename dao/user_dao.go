package dao

import (
	"github.com/gocql/gocql"
	"github.com/twinj/uuid"
	"log"
	core "peeple/areyouin/common"
)

const (
	MAX_NUM_FRIENDS = 1000
	GRACE_PERIOD_MS = 20 * 1000 // 20s
)

type UserDAO struct {
	session *gocql.Session
}

func NewUserDAO(session *gocql.Session) core.UserDAO {
	return &UserDAO{session: session}
}

// Insert a new user into Cassandra involving tables user_account, user_email_credentials
// and user_facebook_credentials. It takes account race conditions like two users trying
// to create the same account simultaneously. If user email is already used this operation
// fails. In the same way, if FB account is already linked to a valid user account (row
// exists in user_account and user_email_credential), then operation fails. With this
// implementation, orphan fb rows are removed after a grace period time in order to enable
// retry logic in the client when, by some reason, registration failed at first attempt.
// In contrast to the older implementation, this new one do not remove user account, it
// only updates orphaned user_facebook_credentials rows that doesn't point to a valid
// user account. In order to work properly, a row in user_email_credentials must be
// inserted before GRACE_PERIOD_MS. Otherwise, another account could reclaim that FB
// for itself with another Insert call.
func (dao *UserDAO) Insert(user *core.UserAccount) error {

	dao.checkSession()

	// Check if user account has a valid ID and email or fb credentials
	if _, err := user.IsValid(); err != nil {
		return err
	}

	// 1) Check if the given e-mail already exists
	if exist, err := dao.ExistEmail(user.Email); exist {
		return ErrEmailAlreadyExists
	} else if err != nil {
		return err
	}

	// E-mail doesn't exist. Then continue
	insert_state := 0

	// Clean logic
	defer func() {
		switch insert_state {
		case 1:
			dao.DeleteUserAccount(user.Id)
		case 2:
			dao.DeleteFacebookCredentials(user.Fbid)
			dao.DeleteUserAccount(user.Id) // Delete always last
		}
	}()

	// 2) Insert into user_account
	if _, err := dao.insertUserAccount(user); err != nil {
		return err
	}

	// 3) Try to insert Facebook credentials considering collisions.
	// See insertFacebookCredentials for more info. If two users try to insert the same
	// FbId, only one of them will succeed.
	insert_state = 1

	if user.HasFacebookCredentials() {
		if _, err := dao.insertFacebookCredentials(user.Fbid, user.Fbtoken, user.Id); err != nil {
			return err
		}
	}

	// 4) Finally, insert e-mail into user_email_credentials. This insert is the most
	// important because it makes the valid the user account warrantying that user_account
	// and user_facebook_credentials also exist. If two users reach this point simultaneously
	// then only one of them will succeed and the other one will fail.
	insert_state = 3 // assume it gonna succeed

	if user.HasEmailCredentials() {
		if _, err := dao.insertEmailCredentials(user.Id, user.Email, user.Password); err != nil {
			insert_state = 2
			return err
		}
	} else if _, err := dao.insertEmail(user.Id, user.Email); err != nil {
		insert_state = 2
		return err
	}

	return nil
}

// Returns true if given e-mail exists, otherwise it returns false.
func (dao *UserDAO) ExistEmail(email string) (bool, error) {

	dao.checkSession()

	stmt := `SELECT user_id FROM user_email_credentials WHERE email = ? LIMIT 1`

	if err := dao.session.Query(stmt, email).Scan(nil); err != nil {
		if err == ErrNotFound {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

// Check if a user exists. Returns true if exists and false if it doesn't.
// If an error happend returns (false, error)
/*func (dao *UserDAO) ExistsUserAccount(user_id uint64) (bool, error) {

	dao.checkSession()

	stmt := `SELECT user_id FROM user_account WHERE user_id = ? LIMIT 1`

	if err := dao.session.Query(stmt, user_id).Scan(nil); err != nil {
		if err == ErrNotFound {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}*/

// Returns an user id corresponding to the given e-mail. If it doesn't exist
// or an error happens, returns (0, error).
func (dao *UserDAO) GetIDByEmail(email string) (uint64, error) {

	dao.checkSession()

	stmt := `SELECT user_id FROM user_email_credentials WHERE email = ? LIMIT 1`
	var user_id uint64

	if err := dao.session.Query(stmt, email).Scan(&user_id); err != nil {
		return 0, err
	}

	return user_id, nil
}

// Returns all of the IDs associated with the given fb_id. If no id is foound
// an empty slide is returned
func (dao *UserDAO) GetIDByFacebookID(fb_id string) (uint64, error) {

	dao.checkSession()

	stmt := `SELECT user_id FROM user_facebook_credentials WHERE fb_id = ?`
	var user_id uint64

	if err := dao.session.Query(stmt, fb_id).Scan(&user_id); err != nil {
		return 0, err
	}

	return user_id, nil
}

func (dao *UserDAO) LoadEmailCredential(email string) (credent *core.EmailCredential, err error) {

	dao.checkSession()

	if email == "" {
		return nil, ErrInvalidEmail
	}

	stmt := `SELECT password, salt, user_id FROM user_email_credentials WHERE email = ? LIMIT 1`
	q := dao.session.Query(stmt, email)

	var pass_slice, salt_slice []byte
	var uid uint64

	// FIXME: Scan doesn't work with array[:] notation
	if err = q.Scan(&pass_slice, &salt_slice, &uid); err == nil {

		credent = &core.EmailCredential{
			Email:  email,
			UserId: uid,
		}

		// HACK: Copy slices to vectors
		copy(credent.Password[:], pass_slice)
		copy(credent.Salt[:], salt_slice)
		// --- End Hack ---
	}

	return
}

func (dao *UserDAO) LoadFacebookCredential(fbid string) (credent *core.FacebookCredential, err error) {

	dao.checkSession()

	stmt := `SELECT fb_token, user_id, created_date FROM user_facebook_credentials WHERE fb_id = ? LIMIT 1`
	q := dao.session.Query(stmt, fbid)

	var fb_token string
	var uid uint64
	var created_date int64

	if err = q.Scan(&fb_token, &uid, &created_date); err == nil {

		credent = &core.FacebookCredential{
			Fbid:        fbid,
			Fbtoken:     fb_token,
			UserId:      uid,
			CreatedDate: created_date,
		}
	}

	return
}

// Checks if the given user_id belongs to an existing and valid account. Returns true
// if the account is valid or false otherwise. If account isn't found or something
// unexpected happens, it returns also an error.
func (dao *UserDAO) CheckValidAccount(user_id uint64, check_credentials bool) (bool, error) {

	stmt_user := `SELECT email, fb_id FROM user_account WHERE user_id = ? LIMIT 1`
	query_user := dao.session.Query(stmt_user, user_id)
	var email string
	var fb_id string

	if err := query_user.Scan(&email, &fb_id); err != nil || email == "" {
		return false, err
	}

	email_credential, err := dao.LoadEmailCredential(email)

	if err == ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if user_id != email_credential.UserId {
		return false, nil
	}

	if check_credentials {

		// First, check e-mail credential
		if email_credential.Password != core.EMPTY_ARRAY_32B && email_credential.Salt != core.EMPTY_ARRAY_32B {
			return true, nil
		}

		// If e-mail credential isn't set, check Facebook
		if fb_credential, err := dao.LoadFacebookCredential(fb_id); err == nil && user_id == fb_credential.UserId {
			return true, nil
		} else if err == ErrNotFound {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

// Returns if given crentials for the given user are valid or not. If credential does not exist it
// consider that are invalid. So, no ErrNotFound is returned.
func (dao *UserDAO) CheckValidCredentials(user_id uint64, email string, fb_id string) (bool, error) {

	email_credential, err := dao.LoadEmailCredential(email)

	if err == ErrNotFound || err == ErrInvalidEmail {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if user_id != email_credential.UserId {
		return false, nil
	}

	// First, check e-mail credential
	if email_credential.Password != core.EMPTY_ARRAY_32B && email_credential.Salt != core.EMPTY_ARRAY_32B {
		return true, nil
	}

	// If e-mail credential isn't set, check Facebook
	if fb_credential, err := dao.LoadFacebookCredential(fb_id); err == nil && user_id == fb_credential.UserId {
		return true, nil
	} else if err == ErrNotFound {
		return false, nil
	} else {
		return false, err
	}
}

// Check if exists an user with given e-mail and password. Returns user id if
// succeed or 0 if doesn't.
func (dao *UserDAO) CheckEmailCredentials(email string, password string) (user_id uint64, err error) {

	dao.checkSession()

	stmt := `SELECT password, salt, user_id FROM user_email_credentials WHERE email = ? LIMIT 1`
	q := dao.session.Query(stmt, email)

	var pass_slice, salt_slice []byte
	var uid uint64
	user_id = 0

	// FIXME: Scan doesn't work with array[:] notation
	if err = q.Scan(&pass_slice, &salt_slice, &uid); err == nil {

		// HACK: Copy slices to vectors
		var pass, salt [32]byte
		copy(pass[:], pass_slice)
		copy(salt[:], salt_slice)
		// --- End Hack ---

		hashedPassword := core.HashPasswordWithSalt(password, salt)
		if hashedPassword == pass {
			user_id = uid
		}
	} else if err != gocql.ErrNotFound {
		log.Println("CheckEmailCredentials:", err)
	}

	return
}

func (dao *UserDAO) CheckAuthToken(user_id uint64, auth_token string) (bool, error) {

	dao.checkSession()

	stmt := `SELECT auth_token FROM user_account WHERE user_id = ? LIMIT 1`
	q := dao.session.Query(stmt, user_id)

	var stored_token gocql.UUID

	err := q.Scan(&stored_token)

	if err == gocql.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if auth_token != stored_token.String() {
		return false, nil
	}

	return true, nil
}

func (dao *UserDAO) SetAuthToken(user_id uint64, auth_token uuid.UUID) error {

	dao.checkSession()

	stmt := `UPDATE user_account SET auth_token = ?
						WHERE user_id = ?`

	return dao.session.Query(stmt, auth_token.String(), user_id).Exec()
}

func (dao *UserDAO) SetLastConnection(user_id uint64, time int64) error {

	dao.checkSession()

	stmt := `UPDATE user_account SET last_connection = ?
						WHERE user_id = ?`

	return dao.session.Query(stmt, time, user_id).Exec()
}

func (dao *UserDAO) SetFacebookAccessToken(user_id uint64, fb_id string, fb_token string) error {

	dao.checkSession()

	batch := dao.session.NewBatch(gocql.LoggedBatch) // the primary use case of a logged batch is when you need to keep tables in sync with one another, and NOT performance.

	batch.Query(`UPDATE user_facebook_credentials SET fb_token = ? WHERE fb_id = ?`,
		fb_token, fb_id)

	batch.Query(`UPDATE user_account SET fb_id = ?, fb_token = ? WHERE user_id = ?`,
		fb_id, fb_token, user_id)

	return dao.session.ExecuteBatch(batch)
}

func (dao *UserDAO) SetAuthTokenAndFBToken(user_id uint64, auth_token uuid.UUID, fb_id string, fb_token string) error {

	dao.checkSession()

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(`UPDATE user_facebook_credentials SET fb_token = ? WHERE fb_id = ?`,
		fb_token, fb_id)

	batch.Query(`UPDATE user_account SET auth_token = ?, fb_id = ?, fb_token = ? WHERE user_id = ?`,
		auth_token.String(), fb_id, fb_token, user_id)

	return dao.session.ExecuteBatch(batch)
}

func (dao *UserDAO) SetIIDToken(user_id uint64, iid_token string) error {

	dao.checkSession()

	stmt := `UPDATE user_account SET iid_token = ?
						WHERE user_id = ?`

	return dao.session.Query(stmt, iid_token, user_id).Exec()
}

func (dao *UserDAO) LoadAllUsers() ([]*core.UserAccount, error) {

	dao.checkSession()

	stmt := `SELECT user_id, auth_token, email, email_verified, name, fb_id, fb_token,
						iid_token, last_connection, created_date, picture_digest
						FROM user_account`

	iter := dao.session.Query(stmt).Iter()

	if iter == nil {
		return nil, ErrNilPointer
	}

	users := make([]*core.UserAccount, 0, 1000)
	var user_id uint64
	var auth_token gocql.UUID
	var email string
	var email_verified bool
	var name string
	var fbid string
	var fbtoken string
	var iidtoken string
	var last_connection int64
	var created int64
	var digest []byte

	for iter.Scan(&user_id, &auth_token, &email, &email_verified, &name,
		&fbid, &fbtoken, &iidtoken, &last_connection, &created, &digest) {
		user := &core.UserAccount{
			Id:             user_id,
			AuthToken:      uuid.New(auth_token.Bytes()),
			Email:          email,
			EmailVerified:  email_verified,
			Name:           name,
			Fbid:           fbid,
			Fbtoken:        fbtoken,
			IIDtoken:       iidtoken,
			LastConnection: last_connection,
			CreatedDate:    created,
			PictureDigest:  digest,
		}
		users = append(users, user)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return users, nil
}

func (dao *UserDAO) Load(user_id uint64) (*core.UserAccount, error) {

	dao.checkSession()

	stmt := `SELECT user_id, auth_token, email, email_verified, name, fb_id, fb_token,
						iid_token, last_connection, created_date, picture_digest
						FROM user_account
						WHERE user_id = ? LIMIT 1`

	q := dao.session.Query(stmt, user_id)

	user := core.NewEmptyUserAccount()
	var auth_token gocql.UUID

	err := q.Scan(&user.Id, &auth_token, &user.Email, &user.EmailVerified, &user.Name,
		&user.Fbid, &user.Fbtoken, &user.IIDtoken, &user.LastConnection, &user.CreatedDate,
		&user.PictureDigest)

	if err != nil {
		log.Println("UserDAO Load:", err)
		return nil, err
	}

	user.AuthToken = uuid.New(auth_token.Bytes())

	return user, nil
}

func (dao *UserDAO) LoadWithPicture(user_id uint64) (*core.UserAccount, error) {

	dao.checkSession()

	stmt := `SELECT user_id, auth_token, email, email_verified, name, fb_id, fb_token,
						iid_token, last_connection, created_date, profile_picture, picture_digest
						FROM user_account
						WHERE user_id = ? LIMIT 1`

	q := dao.session.Query(stmt, user_id)

	user := core.NewEmptyUserAccount()
	var auth_token gocql.UUID

	err := q.Scan(&user.Id, &auth_token, &user.Email, &user.EmailVerified, &user.Name,
		&user.Fbid, &user.Fbtoken, &user.IIDtoken, &user.LastConnection, &user.CreatedDate,
		&user.Picture, &user.PictureDigest)

	if err != nil {
		log.Println("UserDAO Load:", err)
		return nil, err
	}

	user.AuthToken = uuid.New(auth_token.Bytes())

	return user, nil
}

func (dao *UserDAO) LoadByEmail(email string) (*core.UserAccount, error) {

	dao.checkSession()

	user_id, err := dao.GetIDByEmail(email)
	if err != nil {
		return nil, err
	}
	return dao.Load(user_id)
}

func (dao *UserDAO) SaveProfilePicture(user_id uint64, picture *core.Picture) error {
	dao.checkSession()
	stmt := `UPDATE user_account SET profile_picture = ?, picture_digest = ?
						WHERE user_id = ?`
	return dao.session.Query(stmt, picture.RawData, picture.Digest, user_id).Exec()
}

func (dao *UserDAO) insertUserAccount(user *core.UserAccount) (ok bool, err error) {

	dao.checkSession()

	var query *gocql.Query

	if user.HasFacebookCredentials() {

		insertUserAccount := `INSERT INTO	user_account
			(user_id, auth_token, email, email_verified, name, fb_id, fb_token, created_date)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			IF NOT EXISTS`

		query = dao.session.Query(insertUserAccount,
			user.Id,
			user.AuthToken.String(),
			user.Email,
			user.EmailVerified,
			user.Name,
			user.Fbid,
			user.Fbtoken,
			user.CreatedDate)
	} else {

		insertUserAccount := `INSERT INTO	user_account
			(user_id, auth_token, email, email_verified, name, created_date)
			VALUES (?, ?, ?, ?, ?, ?)
			IF NOT EXISTS`

		query = dao.session.Query(insertUserAccount,
			user.Id,
			user.AuthToken.String(),
			user.Email,
			user.EmailVerified,
			user.Name,
			user.CreatedDate)
	}

	return query.ScanCAS(nil)
}

func (dao *UserDAO) insertEmailCredentials(user_id uint64, email string, password string) (ok bool, err error) {

	dao.checkSession()

	if email == "" || password == "" || user_id == 0 {
		return false, ErrInvalidArg
	}

	// Hash Password
	salt, err := core.NewRandomSalt32()
	if err != nil {
		return false, ErrUnexpected
	}

	hashedPassword := core.HashPasswordWithSalt(password, salt)

	insertUserEmailCredentials := `INSERT INTO user_email_credentials
			(email, user_id, password, salt)
			VALUES (?, ?, ?, ?)
			IF NOT EXISTS`

	return dao.session.Query(insertUserEmailCredentials, email, user_id,
		hashedPassword[:], salt[:]).ScanCAS(nil)
}

func (dao *UserDAO) insertEmail(user_id uint64, email string) (ok bool, err error) {

	dao.checkSession()

	if email == "" || user_id == 0 {
		return false, ErrInvalidArg
	}

	insertUserEmail := `INSERT INTO user_email_credentials
			(email, user_id)
			VALUES (?, ?)
			IF NOT EXISTS`

	return dao.session.Query(insertUserEmail, email, user_id).ScanCAS(nil)
}

// Try to insert Facebook credentials. If it fails because of a collision, retrieve
// the row causing that collision, compare created_date with grace_period and check if
// that row belongs to a valid account. If grace_period seconds have elapsed since
// created_date and account isn't valid, then remove row causing conflict and retry.
// Otherwise, if account is valid, returns ErrFacebookAlreadyExists. If grace_period
// seconds haven't elapsed yet since created_date then return ErrGracePeriod .
func (dao *UserDAO) insertFacebookCredentials(fb_id string, fb_token string, user_id uint64) (ok bool, err error) {

	dao.checkSession()

	if fb_id == "" || fb_token == "" || user_id == 0 {
		return false, ErrInvalidArg
	}

	insert_stmt := `INSERT INTO user_facebook_credentials (fb_id, fb_token, user_id, created_date)
		VALUES (?, ?, ?, ?) IF NOT EXISTS`

	current_date := core.GetCurrentTimeMillis()
	query_insert := dao.session.Query(insert_stmt, fb_id, fb_token, user_id, current_date)

	var old_fbid string
	var old_token string
	var old_uid uint64
	var created_date int64

	applied, err := query_insert.ScanCAS(&old_fbid, &created_date, &old_token, &old_uid)
	if err != nil {
		return false, err
	}

	// Retry logic
	if !applied {
		// Grace period expired. Check if account is valid. If not, overwrite row
		if (created_date + GRACE_PERIOD_MS) < current_date {
			if valid, err := dao.CheckValidAccount(old_uid, false); err != nil && err != ErrNotFound {
				return false, err // error happened
			} else if valid {
				return false, ErrFacebookAlreadyExists
			} else { // is invalid account or not exist
				update_stmt := `UPDATE user_facebook_credentials SET fb_token = ?, user_id = ?, created_date = ?
					WHERE fb_id = ? IF created_date < ?`
				current_date = core.GetCurrentTimeMillis()
				query_update := dao.session.Query(update_stmt, fb_token, user_id, current_date, fb_id, current_date-GRACE_PERIOD_MS)
				if applied, err = query_update.ScanCAS(nil); err != nil {
					return false, err
				} else if !applied {
					return false, ErrGracePeriod
				}
			}
		} else {
			return false, ErrGracePeriod
		}
	}

	return applied, err // returns true, nil
}

// User information is spread in three tables: user_account, user_email_credentials
// and user_facebook_credentials. So, in order to delete a user, it's needed an
// user_id, e-mail and, likely, a Facebook ID
// FIXME: I should also delete this user to all of their friends
func (dao *UserDAO) Delete(user *core.UserAccount) error {

	dao.checkSession()

	friendDAO := NewFriendDAO(dao.session)

	friends, err := friendDAO.LoadFriends(user.Id, 0)
	if err != nil {
		return err
	}

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	// Delete email_credential
	batch.Query(`DELETE FROM user_email_credentials WHERE email = ?`, user.Email) // Now account is invalid

	// Delete user from his user's friends
	for _, friend := range friends {
		batch.Query(`DELETE FROM user_friends WHERE user_id = ? AND group_id = ? AND friend_id = ?`,
			friend.UserId, 0, user.Id)
	}

	// Delete user friends
	batch.Query(`DELETE FROM user_friends WHERE user_id = ? AND group_id = ?`, user.Id, 0) // Always after delete itself from other friends

	// Delete Facebook credential
	if user.HasFacebookCredentials() {
		batch.Query(`DELETE FROM user_facebook_credentials WHERE fb_id = ?`, user.Fbid)
	}

	// Delete Thumbnails
	batch.Query(`DELETE FROM thumbnails WHERE id = ?`, user.Id)

	// Delete account
	batch.Query(`DELETE FROM user_account WHERE user_id = ?`, user.Id) // Always last

	return dao.session.ExecuteBatch(batch)
}

func (dao *UserDAO) DeleteUserAccount(user_id uint64) error {
	dao.checkSession()
	if user_id == 0 {
		return ErrInvalidArg
	}
	return dao.session.Query(`DELETE FROM user_account WHERE user_id = ?`, user_id).Exec()
}

func (dao *UserDAO) DeleteEmailCredentials(email string) error {
	dao.checkSession()
	if email == "" {
		return ErrInvalidArg
	}
	return dao.session.Query(`DELETE FROM user_email_credentials WHERE email = ?`, email).Exec()
}

func (dao *UserDAO) DeleteFacebookCredentials(fb_id string) error {
	dao.checkSession()
	if fb_id == "" {
		return ErrInvalidArg
	}
	return dao.session.Query(`DELETE FROM user_facebook_credentials	WHERE fb_id = ?`, fb_id).Exec()
}

func (dao *UserDAO) checkSession() {
	if dao.session == nil {
		panic(ErrNoSession)
	}
}
