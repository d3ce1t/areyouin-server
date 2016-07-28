package cqldao

import (
	"peeple/areyouin/api"
	"peeple/areyouin/utils"

	"github.com/gocql/gocql"
)

var (
	EMPTY_ARRAY_32B = [32]byte{}
)

func (d *UserDAO) loadUserAccount(userId int64) (*api.UserDTO, error) {

	checkSession(d.session)

	if userId == 0 {
		return nil, api.ErrNotFound
	}

	stmt := `SELECT user_id, auth_token, email, email_verified, name, fb_id, fb_token,
						iid_token, network_version, last_connection, created_date, picture_digest
						FROM user_account
						WHERE user_id = ? LIMIT 1`

	q := d.session.Query(stmt, userId)

	dto := new(api.UserDTO)

	err := q.Scan(&dto.Id, &dto.AuthToken, &dto.Email, &dto.EmailVerified, &dto.Name,
		&dto.FbId, &dto.FbToken, &dto.IidToken.Token, &dto.IidToken.Version, &dto.LastConn,
		&dto.CreatedDate, &dto.PictureDigest)

	if err != nil {
		return nil, convErr(err)
	}

	return dto, nil
}

func (d *UserDAO) loadEmailCredential(email string) (*emailCredential, error) {

	checkSession(d.session)

	if email == "" {
		return nil, api.ErrNotFound
	}

	stmt := `SELECT password, salt, user_id FROM user_email_credentials WHERE email = ? LIMIT 1`
	q := d.session.Query(stmt, email)

	var pass_slice, salt_slice []byte
	var uid int64

	// FIXME: Scan doesn't work with array[:] notation
	if err := q.Scan(&pass_slice, &salt_slice, &uid); err != nil {
		return nil, convErr(err)
	}

	credent := &emailCredential{
		Email:  email,
		UserId: uid,
	}

	// HACK: Copy slices to vectors
	copy(credent.Password[:], pass_slice)
	copy(credent.Salt[:], salt_slice)
	// --- End Hack ---

	return credent, nil
}

func (d *UserDAO) loadFacebookCredential(fbId string) (*fbCredential, error) {

	checkSession(d.session)

	if fbId == "" {
		return nil, api.ErrInvalidArg
	}

	stmt := `SELECT fb_token, user_id, created_date FROM user_facebook_credentials WHERE fb_id = ? LIMIT 1`
	q := d.session.Query(stmt, fbId)

	cred := new(fbCredential)

	if err := q.Scan(&cred.FbToken, &cred.UserId, &cred.CreatedDate); err != nil {
		return nil, convErr(err)
	}

	cred.FbId = fbId

	return cred, nil
}

// Returns an user id corresponding to the given e-mail. If it doesn't exist
// or an error happens, returns (0, error).
func (d *UserDAO) getIDByEmail(email string) (int64, error) {

	checkSession(d.session)

	if email == "" {
		return 0, api.ErrNotFound
	}

	stmt := `SELECT user_id FROM user_email_credentials WHERE email = ? LIMIT 1`
	var user_id int64

	if err := d.session.Query(stmt, email).Scan(&user_id); err != nil {
		return 0, convErr(err)
	}

	return user_id, nil
}

func (d *UserDAO) getIDByFacebookID(fb_id string) (int64, error) {

	checkSession(d.session)

	if fb_id == "" {
		return 0, api.ErrNotFound
	}

	stmt := `SELECT user_id FROM user_facebook_credentials WHERE fb_id = ?`
	var user_id int64

	if err := d.session.Query(stmt, fb_id).Scan(&user_id); err != nil {
		return 0, convErr(err)
	}

	return user_id, nil
}

// Returns true if given e-mail exists, otherwise it returns false.
func (d *UserDAO) existEmail(email string) (bool, error) {

	checkSession(d.session)

	if _, err := d.getIDByEmail(email); err == api.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func (dao *UserDAO) insertUserAccount(user *api.UserDTO) (bool, error) {

	checkSession(dao.session)

	var query *gocql.Query

	if user.FbId != "" && user.FbToken != "" {

		insertUserAccount := `INSERT INTO	user_account
			(user_id, auth_token, email, email_verified, name, fb_id, fb_token, created_date)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			IF NOT EXISTS`

		query = dao.session.Query(insertUserAccount, user.Id, user.AuthToken, user.Email,
			user.EmailVerified, user.Name, user.FbId, user.FbToken, user.CreatedDate)
	} else {

		insertUserAccount := `INSERT INTO	user_account
			(user_id, auth_token, email, email_verified, name, created_date)
			VALUES (?, ?, ?, ?, ?, ?)
			IF NOT EXISTS`

		query = dao.session.Query(insertUserAccount, user.Id, user.AuthToken,
			user.Email, user.EmailVerified, user.Name, user.CreatedDate)
	}

	ok, err := query.ScanCAS(nil)

	return ok, convErr(err)
}

// Insert email credentials into db. If a credential with same e-mail address
// already exists, an error will be triggered.
func (d *UserDAO) insertEmailCredentials(cred *emailCredential) (bool, error) {

	checkSession(d.session)

	if cred == nil || cred.UserId == 0 || cred.Email == "" ||
		cred.Password == EMPTY_ARRAY_32B || cred.Salt == EMPTY_ARRAY_32B {
		return false, api.ErrInvalidArg
	}

	insertEmailCredentials := `INSERT INTO user_email_credentials
			(email, user_id, password, salt)
			VALUES (?, ?, ?, ?)
			IF NOT EXISTS`

	ok, err := d.session.Query(insertEmailCredentials, cred.Email, cred.UserId,
		cred.Password[:], cred.Salt[:]).ScanCAS(nil)

	return ok, convErr(err)
}

func (d *UserDAO) insertEmail(user_id int64, email string) (bool, error) {

	checkSession(d.session)

	if email == "" || user_id == 0 {
		return false, api.ErrInvalidArg
	}

	insertUserEmail := `INSERT INTO user_email_credentials
			(email, user_id)
			VALUES (?, ?)
			IF NOT EXISTS`

	ok, err := d.session.Query(insertUserEmail, email, user_id).ScanCAS(nil)

	return ok, convErr(err)
}

// Try to insert Facebook credentials. If it fails because of a collision, retrieve
// the row causing that collision, compare created_date with grace_period and check if
// that row belongs to a valid account. If grace_period seconds have elapsed since
// created_date and account isn't valid, then remove row causing conflict and retry.
// Otherwise, if account is valid, returns ErrFacebookAlreadyExists. If grace_period
// seconds haven't elapsed yet since created_date then return ErrGracePeriod .
func (d *UserDAO) insertFacebookCredentials(userId int64, fbId string, fbToken string) (ok bool, err error) {

	checkSession(d.session)

	if fbId == "" || fbToken == "" || userId == 0 {
		return false, api.ErrInvalidArg
	}

	insert_stmt := `INSERT INTO user_facebook_credentials (fb_id, fb_token, user_id, created_date)
		VALUES (?, ?, ?, ?) IF NOT EXISTS`

	currentDate := utils.GetCurrentTimeMillis()
	query_insert := d.session.Query(insert_stmt, fbId, fbToken, userId, currentDate)

	var old_fbid string
	var old_token string
	var old_uid int64
	var created_date int64

	// TODO: Test order!
	applied, err := query_insert.ScanCAS(&old_fbid, &created_date, &old_token, &old_uid)
	if err != nil {
		return false, convErr(err)
	}

	if !applied {

		// Retry logic

		if (created_date + GRACE_PERIOD_MS) < currentDate {

			// Grace period expired. Check if account is valid. If it isn't, overwrite row

			if _, err := d.LoadByFB(old_fbid); err == ErrInconsistency || err == api.ErrNotFound {

				// Account doesn't exist or is invalid (no e-mail credential)

				update_stmt := `UPDATE user_facebook_credentials SET fb_token = ?, user_id = ?, created_date = ?
					WHERE fb_id = ? IF created_date < ?`
				currentDate = utils.GetCurrentTimeMillis()
				query_update := d.session.Query(update_stmt, fbToken, userId, currentDate,
					fbId, currentDate-GRACE_PERIOD_MS)
				if applied, err = query_update.ScanCAS(nil); err != nil {
					return false, convErr(err)
				} else if !applied {
					return false, ErrGracePeriod
				}

			} else if err != nil {
				return false, err
			} else {
				return false, api.ErrFacebookAlreadyExists
			}

		} else {
			return false, ErrGracePeriod
		}
	}

	return applied, err // returns true, nil
}

func (d *UserDAO) deleteUserAccount(user_id int64) error {
	checkSession(d.session)
	if user_id == 0 {
		return api.ErrInvalidArg
	}
	err := d.session.Query(`DELETE FROM user_account WHERE user_id = ?`, int64(user_id)).Exec()
	return convErr(err)
}

func (d *UserDAO) deleteEmailCredentials(email string) error {
	checkSession(d.session)
	if email == "" {
		return api.ErrInvalidArg
	}
	err := d.session.Query(`DELETE FROM user_email_credentials WHERE email = ?`, email).Exec()
	return convErr(err)
}

func (d *UserDAO) deleteFacebookCredentials(fb_id string) error {
	checkSession(d.session)
	if fb_id == "" {
		return api.ErrInvalidArg
	}
	err := d.session.Query(`DELETE FROM user_facebook_credentials	WHERE fb_id = ?`, fb_id).Exec()
	return convErr(err)
}
