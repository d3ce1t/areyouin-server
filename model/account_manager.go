package model

import (
  "bytes"
  core "peeple/areyouin/common"
  fb "peeple/areyouin/facebook"
  "peeple/areyouin/dao"
  "crypto/sha256"
  "image"
  "errors"
  "log"
  "github.com/twinj/uuid"
)

func newAccountManager(parent *AyiModel, session core.DbSession) *AccountManager {
  return &AccountManager{
    parent: parent,
    dbsession: session,
    userDAO: dao.NewUserDAO(session),
    thumbDAO: dao.NewThumbnailDAO(session),
    friendDAO: dao.NewFriendDAO(session),
  }
}

type AccountManager struct {
  dbsession core.DbSession
  parent *AyiModel
  userDAO core.UserDAO
  thumbDAO core.ThumbnailDAO
  friendDAO core.FriendDAO
}

// Prominent Errors:
// - core.ErrInvalidName
// - core.ErrInvalidEmail
// - core.ErrInvalidPassword
// - core.ErrNoCredentials
// - facebook.ErrFacebookAccessForbidden
func (self *AccountManager) CreateUserAccount(name string, email string, password string, phone string, fbId string, fbToken string) (*core.UserAccount, error) {

  // Create new user account object
  user := core.NewUserAccount(name, email, password, phone, fbId, fbToken)

  // Check if it's a valid user, so the input was correct
  if _, err := user.IsValid(); err != nil {
    return nil, err
  }

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
  if err := self.userDAO.Insert(user); err != nil {
    return nil, err
  }

  return user, nil
}

// Prominent Errors:
// - ErrInvalidUserOrPassword
// - Others (except dao.ErrNotFound)
func (self *AccountManager) NewAuthCredentialByEmailAndPassword(email string, password string) (*core.AuthCredential, error) {

  user_id, err := self.userDAO.GetIDByEmailAndPassword(email, password)
  if err == dao.ErrNotFound {
    return nil, ErrInvalidUserOrPassword
  } else if err != nil {
    return nil, err
  }

  new_auth_token := uuid.NewV4().String()

  if err = self.userDAO.SetAuthToken(user_id, new_auth_token); err != nil {
    return nil, err
  }

  return &core.AuthCredential{UserId: user_id, Token: new_auth_token}, nil
}

// Prominent Errors:
// - fb.ErrFacebookAccessForbidden
// - ErrInvalidUserOrPassword
// - ErrModelInconsistency
// - Others (except dao.ErrNotFound)
func (self *AccountManager) NewAuthCredentialByFacebook(fbId string, fbToken string) (*core.AuthCredential, error) {

  // Use Facebook servers to check if the id and token are valid

  fbsession := fb.NewSession(fbToken)
  if _, err := fb.CheckAccess(fbId, fbsession); err != nil {
    return nil, err
  }

  // Check if Facebook user exists also in AreYouIN, i.e. there is a Fbid
  // pointing to a user id

  user_id, err := self.userDAO.GetIDByFacebookID(fbId)
  if err == dao.ErrNotFound {
    return nil, ErrInvalidUserOrPassword
  } else if err != nil {
    return nil, err
  }

  // Moreover, check if this user.fbid actually exists and match provided fbId

  user, err := self.userDAO.Load(user_id)
  if err == dao.ErrNotFound {
    // TODO: Send e-mail to Admin
    log.Printf("* NEW AUTH_CREDENTIAL BY FACEBOOK WARNING: USER %v NOT FOUND: This means a FbId (%v) points to an AyiId (%v) that does not exist. Admin required.\n", user_id, fbId, user_id)
    return nil, ErrModelInconsistency
  } else if err != nil {
    return nil, err
  }

  if user.Fbid != fbId {
    // TODO: Send e-mail to Admin
    log.Printf("* NEW AUTH_CREDENTIAL BY FACEBOOK WARNING: USER %v FBID MISMATCH: This means a FbId (%v) points to an AyiUser (%v) that does point to another FbId (%v). Admin required.\n", user_id, fbId, user_id, user.Fbid)
    return nil, ErrModelInconsistency
  }

  // Check that account linked to given Facebook ID is valid, i.e. it has user_email_credentials (with or without
  // password). It may happen that a row in user_email_credentials exists but doesn't have a password.
  // Moreover, it may not have Facebook either and it would still be valid. This behaviour is preferred because if
  // this state is found, something had have to be wrong. Under normal conditions, that state should have never
  // happened. So, at this point only existence of e-mail are checked (credentials are ignored).
  // In brief, this only checks that an e-mail exists for this user and points to the his/her corresponding
  // account.

  isValid, err := self.userDAO.CheckValidAccountObject(user.Id, user.Email, user.Fbid, false)
  if err != nil {
    return nil, err
  }

  if !isValid {
    return nil, ErrInvalidUserOrPassword
  }

  new_auth_token := uuid.NewV4().String()

  if err = self.userDAO.SetAuthToken(user_id, new_auth_token); err != nil {
    return nil, err
  }

  return &core.AuthCredential{UserId: user_id, Token: new_auth_token}, nil
}

// Change profile picture in order to let user's friends to see the new picture
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

// Gets AreYouIN users that are friends of user in Facebook
func (self *AccountManager) GetFacebookFriends(user *core.UserAccount) ([]*core.UserAccount, error) {

  // Load Facebook friends that have AreYouIN in Facebook Apps

  fbsession := fb.NewSession(user.Fbtoken)
	fbFriends, err := fb.GetFriends(fbsession)
	if err != nil {
    return nil, errors.New(fb.GetErrorMessage(err))
	}

  // Match Facebook friends to AreYouIN users

  friends := make([]*core.UserAccount, 0, len(fbFriends))

  for _, fbFriend := range fbFriends {

    friend_id, err := self.userDAO.GetIDByFacebookID(fbFriend.Id)
    if err == dao.ErrNotFound {
      // Skip: Facebook user has AreYouIN Facebook App but it's not registered (strangely)
      continue
    } else if err != nil {
			return nil, err
		}

    friendUser, err := self.userDAO.Load(friend_id)
    if err == dao.ErrNotFound {
      // TODO: Send e-mail to Admin
      log.Printf("* GET FACEBOOK FRIENDS WARNING: USER %v NOT FOUND: This means a FbId (%v) points to an AyiId (%v) that does not exist. Admin required.\n", friend_id, fbFriend.Id, friend_id)
      return nil, ErrModelInconsistency
    } else if err != nil {
			return nil, err
		}

    friends = append(friends, friendUser)
  }

  log.Printf("GetFacebookFriends: %v/%v friends found\n", len(friends), len(fbFriends))

  return friends, nil
}

// Adds Facebook friends using AreYouIN to user's friends list
// Returns the list of users that has been added by this operation
func (self *AccountManager) ImportFacebookFriends(user *core.UserAccount) ([]*core.UserAccount, error) {

  // Get areyouin accounts of Facebook friends
  facebookFriends, err := self.GetFacebookFriends(user)
  if err != nil {
    return nil, err
  }

  // Get existing friends as a map
  storedFriends, err := self.friendDAO.LoadFriendsMap(user.GetUserId())
	if err != nil {
		return nil, err
	}

  friends := make([]*core.UserAccount, 0, len(facebookFriends))

	// Filter facebookFriends to get only new friends
	for _, fbFriend := range facebookFriends {

		// Assume that if fbFriend isn't in storedFriends, then user wouldn't be either
		// in the fbFriend friends list
		if _, ok := storedFriends[fbFriend.GetUserId()]; !ok {
      if err := self.friendDAO.MakeFriends(user, fbFriend); err == nil {
        log.Printf("ImportFacebookFriends: %v and %v are now friends\n", user.Id, fbFriend.Id)
        friends = append(friends, fbFriend)
      } else {
        // Log error but do not fail
        log.Printf("ImportFacebookFriends Error (userId=%v, friendId=%v): %v\n", user.Id, fbFriend.Id, err)
        continue
      }
		}
	}

	return friends, nil
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
	err = self.userDAO.SaveProfilePicture(user_id, picture)
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
	err := self.userDAO.SaveProfilePicture(user_id, &core.Picture{RawData: nil, Digest:  nil})
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
