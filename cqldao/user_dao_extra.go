package cqldao

import (
	"github.com/d3ce1t/areyouin-server/api"

	"github.com/gocql/gocql"
)

var (
	EMPTY_ARRAY_32B = [32]byte{}
)

// TODO: Include Password and Salt in user_account table
// TODO: Make up some type of high level iterator in order to control scanning
// steps in model layer.
func (d *UserDAO) Int_LoadAllUserAccount() ([]*api.UserDTO, error) {

	checkSession(d.session)

	stmt := `SELECT user_id, auth_token, email, email_verified, name, fb_id, fb_token,
						iid_token, network_version, platform, last_connection, created_date, picture_digest
						FROM user_account LIMIT 2000`

	iter := d.session.Query(stmt).Iter()

	if iter == nil {
		return nil, ErrUnexpected
	}

	users := make([]*api.UserDTO, 0, 1024)
	var dto api.UserDTO

	for iter.Scan(&dto.Id, &dto.AuthToken, &dto.Email, &dto.EmailVerified, &dto.Name,
		&dto.FbId, &dto.FbToken, &dto.IidToken.Token, &dto.IidToken.Version, &dto.IidToken.Platform,
		&dto.LastConn, &dto.CreatedDate, &dto.PictureDigest) {

		userDTO := new(api.UserDTO)
		*userDTO = dto
		users = append(users, userDTO)
	}

	if err := iter.Close(); err != nil {
		return nil, convErr(err)
	}

	return users, nil
}

func (d *UserDAO) Int_CheckUserConsistency(user *api.UserDTO) (bool, error) {

	// Check Email
	emailCred, err := d.Int_LoadEmailCredential(user.Email)
	if err == api.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	// Check email cred points to this same user id
	if user.Id != emailCred.UserId {
		return false, nil
	}

	if user.FbId != "" {
		// Check Facebook
		fbCred, err := d.Int_LoadFacebookCredential(user.FbId)
		if err == api.ErrNotFound {
			return false, nil
		} else if err != nil {
			return false, err
		}

		// Check fb cred points to this same user id
		if user.Id != fbCred.UserId {
			return false, nil
		}
	}

	// In case password is set, check that it is valid
	emptyPassword := false

	if emailCred.Password == EMPTY_ARRAY_32B && emailCred.Salt == EMPTY_ARRAY_32B {
		emptyPassword = true
	} else if (emailCred.Password != EMPTY_ARRAY_32B && emailCred.Salt == EMPTY_ARRAY_32B) ||
		emailCred.Salt != EMPTY_ARRAY_32B && emailCred.Password == EMPTY_ARRAY_32B {
		return false, nil
	}

	// Check there is at least one credential
	if user.FbId == "" && emptyPassword {
		return false, nil
	}

	return true, nil
}

func (d *UserDAO) Int_LoadUserAccount(userId int64) (*api.UserDTO, error) {

	checkSession(d.session)

	if userId == 0 {
		return nil, api.ErrNotFound
	}

	stmt := `SELECT user_id, auth_token, email, email_verified, name, fb_id, fb_token,
						iid_token, network_version, platform, last_connection, created_date, picture_digest
						FROM user_account
						WHERE user_id = ? LIMIT 1`

	q := d.session.Query(stmt, userId)

	dto := new(api.UserDTO)

	err := q.Scan(&dto.Id, &dto.AuthToken, &dto.Email, &dto.EmailVerified, &dto.Name,
		&dto.FbId, &dto.FbToken, &dto.IidToken.Token, &dto.IidToken.Version, &dto.IidToken.Platform, &dto.LastConn,
		&dto.CreatedDate, &dto.PictureDigest)

	if err != nil {
		return nil, convErr(err)
	}

	return dto, nil
}

func (d *UserDAO) Int_LoadEmailCredential(email string) (*emailCredential, error) {

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

func (d *UserDAO) Int_LoadFacebookCredential(fbId string) (*fbCredential, error) {

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
	err := d.session.Query(`DELETE FROM user_facebook_credentials WHERE fb_id = ?`, fb_id).Exec()
	return convErr(err)
}
