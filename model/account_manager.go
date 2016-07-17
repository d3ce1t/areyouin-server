package model

import (
  "bytes"
  core "peeple/areyouin/common"
  fb "peeple/areyouin/facebook"
  "peeple/areyouin/dao"
  "crypto/sha256"
  "image"
)

func newAccountManager(parent *AyiModel, session core.DbSession) *AccountManager {
  return &AccountManager{parent: parent, dbsession: session}
}

type AccountManager struct {
  dbsession core.DbSession
  parent *AyiModel
}

// core.ErrInvalidName
// core.ErrInvalidEmail
// core.ErrInvalidPassword
// core.ErrNoCredentials
// facebook.ErrFacebookAccessForbidden
func (self *AccountManager) CreateUserAccount(name string, email string, password string, phone string, fbId string, fbToken string) (*core.UserAccount, error) {

  // Create new user account object

  user := core.NewUserAccount(name, email, password, phone, fbId, fbToken)

  // Check if it's a valid user, so the input was correct

  if _, err := user.IsValid(); err != nil {
    return nil, err
  }

  // Try to save into cassandra. Note that UserDAO will check if there is a user
  // that already exists with the same email or fbId

  userDAO := dao.NewUserDAO(self.dbsession)

  // If it's a Facebook account (fbid and fbtoken are not empty) check token
	if user.HasFacebookCredentials() {
		fbsession := fb.NewSession(user.Fbtoken)
		if _, err := fb.CheckAccess(user.Fbid, fbsession); err != nil {
      return nil, err
		}
	}

  // Insert into users database. Insert will fail if an existing user with the same
  // e-mail address already exists, or if the Facebook address is already being used
  // by another user. It also controls orphaned user_facebook_credentials rows due
  // to the way insertion is performed in Cassandra. When orphaned row is found and
  // grace period has not elapsed, an ErrGracePeriod error is triggered. A different
  // error message could be sent to the client whenever this happens. This way client
  // could be notified to wait grace period seconds and retry. However, an OPERATION
  // FAILED message is sent so far. Read UserDAO.insert for more info.
  if err := userDAO.Insert(user); err != nil {
    return nil, err
  }

  return user, nil
}

func (self *AccountManager) ChangeProfilePicture(user *core.UserAccount, picture []byte) error {

  if picture != nil && len(picture) != 0 {

    // Set profile picture

    // Compute digest for picture
    digest := sha256.Sum256(picture)

    corePicture := &core.Picture{
      RawData: picture,
      Digest:  digest[:],
    }

    // Add profile picture
    if err := self.saveProfilePicture(user.GetUserId(), corePicture); err != nil {
      return err
    }

    // TODO: Register UserAccount objects and update fields if needed in order to Keep
    // them updated.
    user.PictureDigest = corePicture.Digest

    if err := self.updateFriendsDigests(user.GetUserId(), corePicture.Digest); err != nil {
      return err
    }

  } else {

    // Remove profile picture

    if err := self.removeProfilePicture(user.GetUserId()); err != nil {
      return err
    }

    user.PictureDigest = nil
  }

  return nil
}

// Saves a profile picture i its original size and alto saves thumbnails for supported dpis
func (self *AccountManager) saveProfilePicture(user_id int64, picture *core.Picture) error {

	// Decode image
	srcImage, _, err := image.Decode(bytes.NewReader(picture.RawData))
	if err != nil {
		return err
	}

	// Check image size is inside bounds
	if srcImage.Bounds().Dx() > core.PROFILE_PICTURE_MAX_WIDTH || srcImage.Bounds().Dy() > core.PROFILE_PICTURE_MAX_HEIGHT {
		return ErrImageOutOfBounds
	}

	// Create thumbnails
	thumbnails, err := core.CreateThumbnails(srcImage, THUMBNAIL_MDPI_SIZE, self.parent.supportedDpi)
	if err != nil {
		return err
	}

	// Save profile picture (max 512x512)
	userDAO := dao.NewUserDAO(self.dbsession)
	err = userDAO.SaveProfilePicture(user_id, picture)
	if err != nil {
		return err
	}

	// Save thumbnails (50x50 to 200x200)
	thumbDAO := dao.NewThumbnailDAO(self.dbsession)
	err = thumbDAO.Insert(user_id, picture.Digest, thumbnails)
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
	userDAO := dao.NewUserDAO(self.dbsession)
	err := userDAO.SaveProfilePicture(user_id, &core.Picture{RawData: nil, Digest:  nil})
	if err != nil {
		return err
	}

	// Remove thumbnails
	thumbDAO := dao.NewThumbnailDAO(self.dbsession)
	err = thumbDAO.Remove(user_id)
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
func (self *AccountManager) updateFriendsDigests(user_id int64, digest []byte) error {

  friendDAO := dao.NewFriendDAO(self.dbsession)
  friends, err := friendDAO.LoadFriends(user_id, ALL_CONTACTS_GROUP)
  if err != nil {
    return err
  }

  for _, friend := range friends {
    err := friendDAO.SetPictureDigest(friend.GetUserId(), user_id, digest)
    if err != nil {
      return err
    }
  }

  return nil
}
