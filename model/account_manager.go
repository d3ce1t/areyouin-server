package model

import (
	"bytes"
	"crypto/sha256"
	"image"
	"peeple/areyouin/api"
	"peeple/areyouin/cqldao"
	fb "peeple/areyouin/facebook"
	"peeple/areyouin/utils"

	"github.com/twinj/uuid"
)

func newAccountManager(parent *AyiModel, session api.DbSession) *AccountManager {
	return &AccountManager{
		parent:         parent,
		dbsession:      session,
		userDAO:        cqldao.NewUserDAO(session),
		thumbDAO:       cqldao.NewThumbnailDAO(session),
		friendDAO:      cqldao.NewFriendDAO(session),
		accessTokenDAO: cqldao.NewAccessTokenDAO(session),
	}
}

type AccountManager struct {
	dbsession      api.DbSession
	parent         *AyiModel
	userDAO        api.UserDAO
	thumbDAO       api.ThumbnailDAO
	friendDAO      api.FriendDAO
	accessTokenDAO api.AccessTokenDAO
}

// Prominent Errors:
// - ErrInvalidName
// - ErrInvalidEmail
// - ErrInvalidPassword
// - ErrNoCredentials
// - facebook.ErrFacebookAccessForbidden
func (self *AccountManager) CreateUserAccount(name string, email string, password string, phone string, fbId string, fbToken string) (*UserAccount, error) {

	// Create new and valid user account object
	user, err := NewUserAccount(name, email, password, phone, fbId, fbToken)
	if err != nil {
		return nil, err
	}

	// If it's a Facebook account (fbid and fbtoken are not empty) check token
	if user.HasFacebook() {
		fbsession := fb.NewSession(user.FbToken())
		if _, err := fb.CheckAccess(user.FbId(), fbsession); err != nil {
			return nil, err
		}
	}

	// Insert into users database
	if err := self.userDAO.Insert(user.AsDTO()); err != nil {
		return nil, err
	}

	return user, nil
}

// Prominent Errors:
// - ErrInvalidUserOrPassword
// - Others (except dao.ErrNotFound)
func (self *AccountManager) NewAuthCredentialByEmailAndPassword(email string, password string) (*AccessToken, error) {

	userDTO, err := self.userDAO.LoadByEmail(email)
	if err == api.ErrNotFound {
		return nil, ErrInvalidUserOrPassword
	} else if err != nil {
		return nil, err
	}

	hashedPassword := utils.HashPasswordWithSalt(password, userDTO.Salt)
	if hashedPassword != userDTO.Password {
		return nil, ErrInvalidUserOrPassword
	}

	// Email and password right. Create a new auth credential

	newAuthToken := uuid.NewV4().String()

	if err = self.userDAO.SetAuthToken(userDTO.Id, newAuthToken); err != nil {
		return nil, err
	}

	return newAccesToken(userDTO.Id, newAuthToken), nil
}

// Prominent Errors:
// - fb.ErrFacebookAccessForbidden
// - ErrInvalidUserOrPassword
// - ErrModelInconsistency
// - Others (except dao.ErrNotFound)
func (self *AccountManager) NewAuthCredentialByFacebook(fbId string, fbToken string) (*AccessToken, error) {

	// Use Facebook servers to check if the id and token are valid

	fbsession := fb.NewSession(fbToken)
	if _, err := fb.CheckAccess(fbId, fbsession); err != nil {
		return nil, err
	}

	// Check if user exists also in AreYouIN

	userDTO, err := self.userDAO.LoadByFB(fbId)
	if err == api.ErrNotFound {
		return nil, ErrInvalidUserOrPassword
	} else if err != nil {
		return nil, err
	}

	newAuthToken := uuid.NewV4().String()

	if err = self.userDAO.SetAuthToken(userDTO.Id, newAuthToken); err != nil {
		return nil, err
	}

	return newAccesToken(userDTO.Id, newAuthToken), nil
}

func (m *AccountManager) NewImageAccessToken(userID int64) (*AccessToken, error) {

	accessToken := newAccesToken(userID, uuid.NewV4().String())

	// Overwrites previous one if exists
	err := m.accessTokenDAO.Insert(accessToken.AsDTO())
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func (self *AccountManager) AuthenticateUser(userId int64, authToken string) (bool, error) {

	user, err := self.userDAO.Load(userId)
	if err == api.ErrNotFound {
		return false, ErrInvalidUserOrPassword
	} else if err != nil {
		return false, err
	}

	if user.AuthToken != "" && user.AuthToken == authToken {
		return true, nil
	} else {
		return false, ErrInvalidUserOrPassword
	}
}

func (self *AccountManager) GetUserAccount(userId int64) (*UserAccount, error) {

	userDTO, err := self.userDAO.Load(userId)
	if err != nil {
		return nil, err
	}

	return newUserFromDTO(userDTO), nil
}

func (self *AccountManager) GetUserAccountByEmail(email string) (*UserAccount, error) {

	userDTO, err := self.userDAO.LoadByEmail(email)
	if err != nil {
		return nil, err
	}

	return newUserFromDTO(userDTO), nil
}

func (self *AccountManager) GetUserAccountByFacebook(fbId string) (*UserAccount, error) {

	userDTO, err := self.userDAO.LoadByFB(fbId)
	if err != nil {
		return nil, err
	}

	return newUserFromDTO(userDTO), nil
}

func (m *AccountManager) ListUsers() ([]*UserAccount, error) {

	usersDTO, err := m.userDAO.LoadAll()
	if err != nil {
		return nil, err
	}

	users := newUserListFromDTO(usersDTO)
	return users, nil
}

func (self *AccountManager) GetPushToken(userId int64) (*IIDToken, error) {

	tokenDTO, err := self.userDAO.LoadIIDToken(userId)
	if err != nil {
		return nil, err
	}

	return newIIDTokenFromDTO(tokenDTO), nil
}

func (m *AccountManager) SetPushToken(userID int64, pushToken *IIDToken) error {

	err := m.userDAO.SetIIDToken(userID, pushToken.AsDTO())
	if err != nil {
		return err
	}

	//user.iidToken = pushToken
	return nil
}

func (self *AccountManager) SetLastConnection(userId int64, time int64) error {
	return self.userDAO.SetLastConnection(userId, time)
}

func (m *AccountManager) ChangePassword(user *UserAccount, newPassword string) error {

	if user == nil {
		return ErrNotFound
	}

	if !isValidPassword(newPassword) {
		return ErrInvalidPassword
	}

	salt := utils.NewRandomSalt32()
	hashedPassword := utils.HashPasswordWithSalt(newPassword, salt)

	if _, err := m.userDAO.SetPassword(user.email, hashedPassword, salt); err != nil {
		return err
	}

	user.emailCred.Password = hashedPassword
	user.emailCred.Salt = salt

	return nil
}

// Change profile picture in order to let user's friends to see the new picture
func (self *AccountManager) ChangeProfilePicture(user *UserAccount, picture []byte) error {

	if picture != nil && len(picture) != 0 {

		// Set profile picture

		// Compute digest for picture
		digest := sha256.Sum256(picture)

		corePicture := &Picture{
			RawData: picture,
			Digest:  digest[:],
		}

		// Add profile picture
		if err := self.saveProfilePicture(user.Id(), corePicture); err != nil {
			return err
		}

		// TODO: Register UserAccount objects and update fields if needed in order to Keep
		// them updated.
		user.pictureDigest = corePicture.Digest

		if err := self.updateFriendsDigests(user.Id(), corePicture.Digest); err != nil {
			return err
		}

	} else {

		// Remove profile picture

		if err := self.removeProfilePicture(user.Id()); err != nil {
			return err
		}

		user.pictureDigest = nil
	}

	return nil
}

/*func (m *AccountManager) DeleteUserAccount(userID int64) error {
	return nil
}*/

// Saves a profile picture i its original size and alto saves thumbnails for supported dpis
func (self *AccountManager) saveProfilePicture(user_id int64, picture *Picture) error {

	// Decode image
	srcImage, _, err := image.Decode(bytes.NewReader(picture.RawData))
	if err != nil {
		return err
	}

	// Check image size is inside bounds
	if srcImage.Bounds().Dx() > PROFILE_PICTURE_MAX_WIDTH || srcImage.Bounds().Dy() > PROFILE_PICTURE_MAX_HEIGHT {
		return ErrImageOutOfBounds
	}

	// Create thumbnails
	thumbnails, err := utils.CreateThumbnails(srcImage, THUMBNAIL_MDPI_SIZE, self.parent.supportedDpi)
	if err != nil {
		return err
	}

	// Save profile picture (max 512x512)
	err = self.userDAO.SaveProfilePicture(user_id, picture.AsDTO())
	if err != nil {
		return err
	}

	// Save thumbnails (50x50 to 200x200)
	err = self.thumbDAO.Insert(user_id, picture.Digest, thumbnails)
	if err != nil {
		return err
	}

	// Update friends' digests
	if err := self.updateFriendsDigests(user_id, picture.Digest); err != nil {
		return err
	}

	return nil
}

func (self *AccountManager) removeProfilePicture(user_id int64) error {

	// Remove profile picture
	emptyPic := &Picture{RawData: nil, Digest: nil}
	err := self.userDAO.SaveProfilePicture(user_id, emptyPic.AsDTO())
	if err != nil {
		return err
	}

	// Remove thumbnails
	err = self.thumbDAO.Remove(user_id)
	if err != nil {
		return err
	}

	// Update friends' digests
	if err := self.updateFriendsDigests(user_id, nil); err != nil {
		return err
	}

	return nil
}

// Store digest in user's friends so that friends can know that user profile picture
// has been changed next time they retrieve user list
func (self *AccountManager) updateFriendsDigests(userId int64, digest []byte) error {

	friends, err := self.friendDAO.LoadFriends(userId, ALL_CONTACTS_GROUP)
	if err != nil {
		return err
	}

	for _, friend := range friends {
		err := self.friendDAO.SetFriendPictureDigest(friend.UserId, userId, digest)
		if err != nil {
			return err
		}
	}

	return nil
}
