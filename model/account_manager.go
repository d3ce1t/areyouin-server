package model

import (
	"bytes"
	"crypto/sha256"
	"github.com/twinj/uuid"
	"image"
	"peeple/areyouin/api"
	"peeple/areyouin/cqldao"
	fb "peeple/areyouin/facebook"
	"peeple/areyouin/utils"
)

func newAccountManager(parent *AyiModel, session api.DbSession) *AccountManager {
	return &AccountManager{
		parent:    parent,
		dbsession: session,
		userDAO:   cqldao.NewUserDAO(session),
		thumbDAO:  cqldao.NewThumbnailDAO(session),
		friendDAO: cqldao.NewFriendDAO(session),
	}
}

type AccountManager struct {
	dbsession api.DbSession
	parent    *AyiModel
	userDAO   api.UserDAO
	thumbDAO  api.ThumbnailDAO
	friendDAO api.FriendDAO
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
func (self *AccountManager) NewAuthCredentialByEmailAndPassword(email string, password string) (*AuthCredential, error) {

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

	new_auth_token := uuid.NewV4().String()

	if err = self.userDAO.SetAuthToken(userDTO.Id, new_auth_token); err != nil {
		return nil, err
	}

	return &AuthCredential{UserId: userDTO.Id, Token: new_auth_token}, nil
}

// Prominent Errors:
// - fb.ErrFacebookAccessForbidden
// - ErrInvalidUserOrPassword
// - ErrModelInconsistency
// - Others (except dao.ErrNotFound)
func (self *AccountManager) NewAuthCredentialByFacebook(fbId string, fbToken string) (*AuthCredential, error) {

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

	new_auth_token := uuid.NewV4().String()

	if err = self.userDAO.SetAuthToken(userDTO.Id, new_auth_token); err != nil {
		return nil, err
	}

	return &AuthCredential{UserId: userDTO.Id, Token: new_auth_token}, nil
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

	return NewUserFromDTO(userDTO), nil
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