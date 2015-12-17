package dao

import (
	core "areyouin/common"
	proto "areyouin/protocol"
	"errors"
	"github.com/gocql/gocql"
	"github.com/twinj/uuid"
	"log"
)

const (
	MAX_NUM_FRIENDS = 1000
)

type UserDAO struct {
	session *gocql.Session
}

func NewUserDAO(session *gocql.Session) core.UserDAO {
	return &UserDAO{session: session}
}

/*
	Check if exists an user with given e-mail and password. Returns user id if
	exists or 0 if doesn't exist.
*/
func (dao *UserDAO) CheckEmailCredentials(email string, password string) uint64 {

	stmt := `SELECT password, salt, user_id FROM user_email_credentials WHERE email = ? LIMIT 1`
	q := dao.session.Query(stmt, email)

	var pass_slice, salt_slice []byte
	var uid uint64

	// FIXME: Scan doesn't work with array[:] notation
	if err := q.Scan(&pass_slice, &salt_slice, &uid); err == nil {

		// HACK: Copy slices to vectors
		var pass, salt [32]byte
		copy(pass[:], pass_slice)
		copy(salt[:], salt_slice)
		// --- End Hack ---

		hashedPassword := core.HashPasswordWithSalt(password, salt)
		if hashedPassword == pass {
			return uid
		}
	} else if err != gocql.ErrNotFound {
		log.Println("CheckEmailCredentials:", err)
	}

	return 0
}

func (dao *UserDAO) CheckAuthToken(user_id uint64, auth_token uuid.UUID) bool {

	stmt := `SELECT auth_token FROM user_account WHERE user_id = ? LIMIT 1`
	q := dao.session.Query(stmt, user_id)

	var stored_token gocql.UUID

	if err := q.Scan(&stored_token); err != nil {
		log.Println("CheckAuthToken:", err)
		return false
	}

	if auth_token.String() != stored_token.String() {
		return false
	}

	return true
}

func (dao *UserDAO) SetAuthToken(user_id uint64, auth_token uuid.UUID) error {

	stmt := `UPDATE user_account SET auth_token = ?
						WHERE user_id = ?`

	return dao.session.Query(stmt, auth_token.String(), user_id).Exec()
}

func (dao *UserDAO) SetLastConnection(user_id uint64, time int64) error {
	stmt := `UPDATE user_account SET last_connection = ?
						WHERE user_id = ?`

	return dao.session.Query(stmt, time, user_id).Exec()
}

func (dao *UserDAO) SetFacebookAccessToken(user_id uint64, fb_id string, fb_token string) error {

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(`UPDATE user_facebook_credentials SET fb_token = ? WHERE fb_id = ?`,
		fb_token, fb_id)

	batch.Query(`UPDATE user_account SET fb_token = ? WHERE user_id = ?`,
		fb_token, user_id)

	return dao.session.ExecuteBatch(batch)
}

func (dao *UserDAO) SetAuthTokenAndFBToken(user_id uint64, auth_token uuid.UUID, fb_id string, fb_token string) error {

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(`UPDATE user_facebook_credentials SET fb_token = ? WHERE fb_id = ?`,
		fb_token, fb_id)

	batch.Query(`UPDATE user_account SET auth_token = ?, fb_token = ? WHERE user_id = ?`,
		auth_token.String(), fb_token, user_id)

	return dao.session.ExecuteBatch(batch)
}

func (dao *UserDAO) GetIDByEmail(email string) uint64 {

	stmt := `SELECT user_id FROM user_email_credentials WHERE email = ? LIMIT 1`
	var user_id uint64

	if err := dao.session.Query(stmt, email).Scan(&user_id); err != nil {
		return 0
	}

	return user_id
}

func (dao *UserDAO) GetIDByFacebookID(fb_id string) uint64 {

	stmt := `SELECT user_id FROM user_facebook_credentials WHERE fb_id = ? LIMIT 1`
	var user_id uint64

	if err := dao.session.Query(stmt, fb_id).Scan(&user_id); err != nil {
		return 0
	}

	return user_id
}

func (dao *UserDAO) Exists(user_id uint64) bool {

	stmt := `SELECT user_id FROM user_account WHERE user_id = ? LIMIT 1`
	exists := false

	if err := dao.session.Query(stmt, user_id).Scan(nil); err == nil {
		exists = true
	}

	return exists
}

func (dao *UserDAO) Load(user_id uint64) *core.UserAccount {

	stmt := `SELECT user_id, auth_token, email, email_verified, name, fb_id, fb_token,
						last_connection, created_date
						FROM user_account
						WHERE user_id = ? LIMIT 1`

	q := dao.session.Query(stmt, user_id)

	user := core.NewEmptyUserAccount()
	var auth_token gocql.UUID

	err := q.Scan(&user.Id, &auth_token, &user.Email, &user.EmailVerified, &user.Name,
		&user.Fbid, &user.Fbtoken, &user.LastConnection, &user.CreatedDate)

	if err != nil {
		log.Println("UserDAO: An error ocurred while loading user account", err)
		return nil
	}

	user.AuthToken = uuid.New(auth_token.Bytes())

	return user
}

func (dao *UserDAO) LoadByEmail(email string) *core.UserAccount {
	user_id := dao.GetIDByEmail(email)
	if user_id == 0 {
		return nil
	}
	return dao.Load(user_id)
}

/*
	Insert a new user into Cassandra involving tables user_account,
	user_email_credentials and user_facebook_credentials.
*/
func (dao *UserDAO) Insert(user *core.UserAccount) (ok bool, err error) {

	// Check if user account has a valid ID and email or fb credentials
	if !core.CheckUserAccount(user) {
		return false, errors.New("UserDAO: Trying to insert an invalid user account")
	}

	// First insert into E-mail credentials to ensure there is no one using the same
	// email address. If there isn't email credentials, there will be Facebook and
	// a valid email address. So try to insert email anyway to check there is no one
	// using it.
	if user.HasEmailCredentials() {
		ok, err = dao.insertEmailCredentials(user.Id, user.Email, user.Password)
		if !ok {
			return
		}
	} else {
		ok, err = dao.insertEmail(user.Id, user.Email)
		if !ok {
			return
		}
	}

	// May also (or only) have Facebook credentials
	if user.HasFacebookCredentials() {
		ok, err = dao.insertFacebookCredentials(user.Id, user.Fbid, user.Fbtoken)
		if !ok {
			dao.DeleteEmailCredentials(user.Email)
			return
		}
	}

	// If this point is reached, then insert user account. UserID must be unique or
	// the operation will fail. Code guarantees that each generated UserID does not
	// collide if each generator running on its own goroutine has a different ID.
	// However, if for any reason the same UserID is generated, the already stored
	// user account does not have to be replaced. In this case, it's needed to
	// manually rollback previously inserted EmailCredentials and/or FacebookCredentials
	ok, err = dao.insertUserAccount(user)
	if err != nil {
		dao.DeleteFacebookCredentials(user.Fbid)
		dao.DeleteEmailCredentials(user.Email)
	}

	return
}

func (dao *UserDAO) insertUserAccount(user *core.UserAccount) (ok bool, err error) {

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

	if email == "" || password == "" || user_id == 0 {
		return false, errors.New("Invalid arguments")
	}

	// Hash Password
	salt, err := core.NewRandomSalt32()
	if err != nil {
		return false, errors.New("insertEmailCredentials error when hashing password")
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

	if email == "" || user_id == 0 {
		return false, errors.New("Invalid arguments")
	}

	insertUserEmail := `INSERT INTO user_email_credentials
			(email, user_id)
			VALUES (?, ?)
			IF NOT EXISTS`

	return dao.session.Query(insertUserEmail, email, user_id).ScanCAS(nil)
}

func (dao *UserDAO) insertFacebookCredentials(user_id uint64, fb_id string, fb_token string) (ok bool, err error) {

	if fb_id == "" || fb_token == "" || user_id == 0 {
		return false, errors.New("Invalid arguments")
	}

	insertFacebookCredentials := `INSERT INTO user_facebook_credentials
		(fb_id, fb_token, user_id)
		VALUES (?, ?, ?)
		IF NOT EXISTS`

	return dao.session.Query(insertFacebookCredentials, fb_id, fb_token, user_id).ScanCAS(nil)
}

/*
	User information is spread in three tables: user_account, user_email_credentials
	and user_facebook_credentials. So, in order to delete a user, it's needed an
	user_id, e-mail and, likely, a Facebook ID
*/
func (dao *UserDAO) Delete(user *core.UserAccount) error {

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(`DELETE FROM user_email_credentials WHERE email = ?`, user.Email)
	batch.Query(`DELETE FROM user_account WHERE user_id = ?`, user.Id)

	if user.HasFacebookCredentials() {
		batch.Query(`DELETE FROM user_facebook_credentials	WHERE fb_id = ?`, user.Fbid)
	}

	return dao.session.ExecuteBatch(batch)
}

func (dao *UserDAO) deleteUserAccount(user_id uint64) error {

	if user_id == 0 {
		return errors.New("Invalid arguments")
	}

	return dao.session.Query(`DELETE FROM user_account WHERE user_id = ?`, user_id).Exec()
}

func (dao *UserDAO) DeleteEmailCredentials(email string) error {

	if email == "" {
		return errors.New("Invalid arguments")
	}

	return dao.session.Query(`DELETE FROM user_email_credentials WHERE email = ?`, email).Exec()
}

func (dao *UserDAO) DeleteFacebookCredentials(fb_id string) error {

	if fb_id == "" {
		return errors.New("Invalid arguments")
	}

	return dao.session.Query(`DELETE FROM user_facebook_credentials	WHERE fb_id = ?`, fb_id).Exec()
}

func (dao *UserDAO) AddFriend(user_id uint64, friend *proto.Friend, group_id int32) error {

	stmt := `INSERT INTO user_friends (user_id, group_id, group_name, friend_id, name)
							VALUES (?, ?, ?, ?, ?)`

	return dao.session.Query(stmt, user_id, group_id, "All Friends", friend.UserId, friend.Name).Exec()
}

func (dao *UserDAO) LoadFriends(user_id uint64, group_id int32) []*proto.Friend {

	stmt := `SELECT friend_id, name FROM user_friends
						WHERE user_id = ? AND group_id = ? LIMIT ?`

	iter := dao.session.Query(stmt, user_id, group_id, MAX_NUM_FRIENDS).Iter()

	friend_list := make([]*proto.Friend, 0, 10)

	var friend_id uint64
	var friend_name string

	for iter.Scan(&friend_id, &friend_name) {
		friend_list = append(friend_list, &proto.Friend{UserId: friend_id, Name: friend_name})
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadFriends:", err)
	}

	return friend_list
}

func (dao *UserDAO) AreFriends(user_id uint64, other_user_id uint64) bool {

	stmt := `SELECT user_id FROM user_friends
		WHERE user_id = ? AND group_id = ? AND friend_id = ?`

	one_way := false
	two_way := false

	if err := dao.session.Query(stmt, user_id, 0, other_user_id).Scan(nil); err == nil { // HACK: 0 group contains ALL_CONTACTS
		one_way = true
		if err := dao.session.Query(stmt, other_user_id, 0, user_id).Scan(nil); err == nil {
			two_way = true
		}
	}

	return one_way && two_way
}
