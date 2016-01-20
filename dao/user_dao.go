package dao

import (
	"github.com/gocql/gocql"
	"github.com/twinj/uuid"
	"log"
	core "peeple/areyouin/common"
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

func (dao *UserDAO) CheckAuthToken(user_id uint64, auth_token uuid.UUID) (bool, error) {

	dao.checkSession()

	stmt := `SELECT auth_token FROM user_account WHERE user_id = ? LIMIT 1`
	q := dao.session.Query(stmt, user_id)

	var stored_token gocql.UUID

	if err_tmp := q.Scan(&stored_token); err_tmp != nil {
		return false, err_tmp
	}

	if auth_token.String() != stored_token.String() {
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

	batch.Query(`UPDATE user_account SET fb_token = ? WHERE user_id = ?`,
		fb_token, user_id)

	return dao.session.ExecuteBatch(batch)
}

func (dao *UserDAO) SetAuthTokenAndFBToken(user_id uint64, auth_token uuid.UUID, fb_id string, fb_token string) error {

	dao.checkSession()

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(`UPDATE user_facebook_credentials SET fb_token = ? WHERE fb_id = ?`,
		fb_token, fb_id)

	batch.Query(`UPDATE user_account SET auth_token = ?, fb_token = ? WHERE user_id = ?`,
		auth_token.String(), fb_token, user_id)

	return dao.session.ExecuteBatch(batch)
}

func (dao *UserDAO) GetIDByEmail(email string) (uint64, error) {

	dao.checkSession()

	stmt := `SELECT user_id FROM user_email_credentials WHERE email = ? LIMIT 1`
	var user_id uint64

	if err := dao.session.Query(stmt, email).Scan(&user_id); err != nil {
		return 0, err
	}

	return user_id, nil
}

func (dao *UserDAO) GetIDByFacebookID(fb_id string) (uint64, error) {

	dao.checkSession()

	stmt := `SELECT user_id FROM user_facebook_credentials WHERE fb_id = ? LIMIT 1`
	var user_id uint64

	if err := dao.session.Query(stmt, fb_id).Scan(&user_id); err != nil {
		return 0, err
	}

	return user_id, nil
}

func (dao *UserDAO) Exists(user_id uint64) (bool, error) {

	dao.checkSession()

	stmt := `SELECT user_id FROM user_account WHERE user_id = ? LIMIT 1`

	if err := dao.session.Query(stmt, user_id).Scan(nil); err != nil {
		return false, err
	}

	return true, nil
}

func (dao *UserDAO) Load(user_id uint64) (*core.UserAccount, error) {

	dao.checkSession()

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
		log.Println("UserDAO.Load:", err)
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

/*
	Insert a new user into Cassandra involving tables user_account,
	user_email_credentials and user_facebook_credentials.
*/
func (dao *UserDAO) Insert(user *core.UserAccount) error {

	dao.checkSession()

	// Check if user account has a valid ID and email or fb credentials
	if !user.IsValid() {
		return ErrInvalidUser
	}

	// First insert into E-mail credentials to ensure there is no one using the same
	// email address. If there isn't email credentials, there will be Facebook and
	// a valid email address. So try to insert email anyway to check there is no one
	// using it.
	if user.HasEmailCredentials() {
		if _, err := dao.insertEmailCredentials(user.Id, user.Email, user.Password); err != nil {
			return err
		}
	} else if _, err := dao.insertEmail(user.Id, user.Email); err != nil {
		return err
	}

	// May also (or only) have Facebook credentials
	if user.HasFacebookCredentials() {
		if _, err := dao.insertFacebookCredentials(user.Id, user.Fbid, user.Fbtoken); err != nil {
			dao.DeleteEmailCredentials(user.Email)
			return err
		}
	}

	// If this point is reached, then insert user account. UserID must be unique or
	// the operation will fail. Code guarantees that each generated UserID does not
	// collide if each generator running on its own goroutine has a different ID.
	// However, if for any reason the same UserID is generated, the already stored
	// user account does not have to be replaced. In this case, it's needed to
	// manually rollback previously inserted EmailCredentials and/or FacebookCredentials
	if _, err := dao.insertUserAccount(user); err != nil {
		dao.DeleteFacebookCredentials(user.Fbid)
		dao.DeleteEmailCredentials(user.Email)
		return err
	}

	return nil
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

func (dao *UserDAO) insertFacebookCredentials(user_id uint64, fb_id string, fb_token string) (ok bool, err error) {

	dao.checkSession()

	if fb_id == "" || fb_token == "" || user_id == 0 {
		return false, ErrInvalidArg
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

	dao.checkSession()

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(`DELETE FROM user_email_credentials WHERE email = ?`, user.Email)
	batch.Query(`DELETE FROM user_account WHERE user_id = ?`, user.Id)
	batch.Query(`DELETE FROM user_friends WHERE user_id = ? AND group_id = ?`, user.Id, 0)
	// FIXME: I should also delete this user to all of their friends

	if user.HasFacebookCredentials() {
		batch.Query(`DELETE FROM user_facebook_credentials WHERE fb_id = ?`, user.Fbid)
	}

	return dao.session.ExecuteBatch(batch)
}

func (dao *UserDAO) deleteUserAccount(user_id uint64) error {

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

func (dao *UserDAO) MakeFriends(user1 *core.Friend, user2 *core.Friend) error {

	dao.checkSession()

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(`INSERT INTO user_friends (user_id, group_id, group_name, friend_id, name)
							VALUES (?, ?, ?, ?, ?)`, user1.UserId, 0, "All Friends", user2.UserId, user2.Name)

	batch.Query(`INSERT INTO user_friends (user_id, group_id, group_name, friend_id, name)
							VALUES (?, ?, ?, ?, ?)`, user2.UserId, 0, "All Friends", user1.UserId, user1.Name)

	return dao.session.ExecuteBatch(batch)
}

func (dao *UserDAO) AddFriend(user_id uint64, friend *core.Friend, group_id int32) error {

	dao.checkSession()

	stmt := `INSERT INTO user_friends (user_id, group_id, group_name, friend_id, name)
							VALUES (?, ?, ?, ?, ?)`

	return dao.session.Query(stmt, user_id, group_id, "All Friends", friend.UserId, friend.Name).Exec()
}

func (dao *UserDAO) DeleteFriendsGroup(user_id uint64, group_id int32) error {

	dao.checkSession()

	if user_id == 0 {
		return ErrInvalidArg
	}

	return dao.session.Query(`DELETE FROM user_friends WHERE user_id = ? AND group_id = ?`,
		user_id, group_id).Exec()
}

func (dao *UserDAO) LoadFriendsIndex(user_id uint64, group_id int32) (map[uint64]*core.Friend, error) {

	dao.checkSession()

	stmt := `SELECT friend_id, name FROM user_friends
						WHERE user_id = ? AND group_id = ? LIMIT ?`

	iter := dao.session.Query(stmt, user_id, group_id, MAX_NUM_FRIENDS).Iter()

	friend_map := make(map[uint64]*core.Friend)

	var friend_id uint64
	var friend_name string

	for iter.Scan(&friend_id, &friend_name) {
		friend_map[friend_id] = &core.Friend{UserId: friend_id, Name: friend_name}
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadFriendsIndex:", err)
		return nil, err
	}

	return friend_map, nil
}

func (dao *UserDAO) LoadFriends(user_id uint64, group_id int32) ([]*core.Friend, error) {

	dao.checkSession()

	stmt := `SELECT friend_id, name FROM user_friends
						WHERE user_id = ? AND group_id = ? LIMIT ?`

	iter := dao.session.Query(stmt, user_id, group_id, MAX_NUM_FRIENDS).Iter()

	friend_list := make([]*core.Friend, 0, 10)

	var friend_id uint64
	var friend_name string

	for iter.Scan(&friend_id, &friend_name) {
		friend_list = append(friend_list, &core.Friend{UserId: friend_id, Name: friend_name})
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadFriends:", err)
		return nil, err
	}

	return friend_list, nil
}

func (dao *UserDAO) AreFriends(user_id uint64, other_user_id uint64) (bool, error) {

	dao.checkSession()

	stmt := `SELECT user_id FROM user_friends
		WHERE user_id = ? AND group_id = ? AND friend_id = ?`

	one_way := false
	two_way := false

	if err := dao.session.Query(stmt, user_id, 0, other_user_id).Scan(nil); err == nil { // HACK: 0 group contains ALL_CONTACTS
		one_way = true
		if err := dao.session.Query(stmt, other_user_id, 0, user_id).Scan(nil); err == nil {
			two_way = true
		} else {
			return false, err
		}
	} else {
		return false, err
	}

	return one_way && two_way, nil
}

// Check if user exists. If user e-mail exists may be orphan due to the way users are
// inserted into cassandra. So it's needed to check if the user related to this e-mail
// also exists. In case it doesn't exist, then delete the e-mail in order to avoid a collision
// when inserting later. Exist
func (dao *UserDAO) ExistWithSanity(user *core.UserAccount) (bool, error) {

	dao.checkSession()

	// Check if e-mail exists
	user_id, err := dao.GetIDByEmail(user.Email)
	if err != nil {
		return false, err
	}

	// If exists, check also if the related user_id also exists
	exist, err := dao.Exists(user_id)
	if err != nil && err != gocql.ErrNotFound {
		return false, err
	}

	if !exist {
		if user.HasFacebookCredentials() {
			if user_id, _ := dao.GetIDByFacebookID(user.Fbid); user_id == user.Id { // FIXME: Errors aren't checked
				dao.DeleteFacebookCredentials(user.Fbid)
			}
		}
		dao.DeleteEmailCredentials(user.Email)
	}

	return exist, nil
}

func (dao *UserDAO) checkSession() {
	if dao.session == nil {
		panic(ErrNoSession)
	}
}
