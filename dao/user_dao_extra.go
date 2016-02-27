package dao

import (
	"github.com/gocql/gocql"
	core "peeple/areyouin/common"
)

func (dao *UserDAO) checkEmailCredentialObject(user_id uint64, credential *core.EmailCredential, ignore_password bool) bool {

	if user_id == 0 || credential == nil {
		return false
	}

	if user_id != credential.UserId {
		return false
	}

	if !ignore_password {
		// First, check e-mail credential
		if credential.Password == core.EMPTY_ARRAY_32B ||
			credential.Salt == core.EMPTY_ARRAY_32B {
			return false
		}
	}

	return true
}

// Check if a user with user_id has a credential with e-mail and password
func (dao *UserDAO) checkEmailCredential(user_id uint64, email string, ignore_password bool) (bool, error) {

	if user_id == 0 || email == "" {
		return false, ErrInvalidArg
	}

	email_credential, err := dao.LoadEmailCredential(email)

	if err == ErrNotFound || err == ErrInvalidEmail {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return dao.checkEmailCredentialObject(user_id, email_credential, ignore_password), nil
}

// Check if a user with user_id has a credential with fb_id
func (dao *UserDAO) checkFacebookCredential(user_id uint64, fb_id string) (bool, error) {

	if user_id == 0 || fb_id == "" {
		return false, ErrInvalidArg
	}

	fb_credential, err := dao.LoadFacebookCredential(fb_id)

	if err == nil && user_id == fb_credential.UserId {
		return true, nil
	} else if err == ErrNotFound {
		return false, nil
	} else {
		return false, err
	}
}

// Returns true if given e-mail exists, otherwise it returns false.
func (dao *UserDAO) existEmail(email string) (bool, error) {

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

// Returns an user id corresponding to the given e-mail. If it doesn't exist
// or an error happens, returns (0, error).
func (dao *UserDAO) getIDByEmail(email string) (uint64, error) {

	dao.checkSession()

	stmt := `SELECT user_id FROM user_email_credentials WHERE email = ? LIMIT 1`
	var user_id uint64

	if err := dao.session.Query(stmt, email).Scan(&user_id); err != nil {
		return 0, err
	}

	return user_id, nil
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

/*func (dao *UserDAO) FindMatchByAuthToken(user_id uint64, auth_token string) (bool, error) {

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
}*/

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

func (dao *UserDAO) checkSession() {
	if dao.session == nil {
		panic(ErrNoSession)
	}
}
