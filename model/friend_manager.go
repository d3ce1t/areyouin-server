package model

import (
	"errors"
	"log"
	"peeple/areyouin/api"
	"peeple/areyouin/cqldao"
	fb "peeple/areyouin/facebook"
)

type FriendManager struct {
	dbsession api.DbSession
	parent    *AyiModel
	userDAO   api.UserDAO
	friendDAO api.FriendDAO
}

func newFriendManager(parent *AyiModel, session api.DbSession) *FriendManager {
	return &FriendManager{
		parent:    parent,
		dbsession: session,
		userDAO:   cqldao.NewUserDAO(session),
		friendDAO: cqldao.NewFriendDAO(session),
	}
}

func (self *FriendManager) LoadAllFriends(user *UserAccount) ([]*Friend, error) {

	friendsDTO, err := self.friendDAO.LoadFriends(user.Id(), ALL_CONTACTS_GROUP)
	if err != nil {
		return nil, err
	}

	friends := make([]*Friend, 0, len(friendsDTO))
	for _, f := range friendsDTO {
		friends = append(friends, NewFriendFromDTO(f))
	}

	return friends, nil
}

// Gets AreYouIN users that are friends of user in Facebook
func (self *FriendManager) GetFacebookFriends(user *UserAccount) ([]*UserAccount, error) {

	// Load Facebook friends that have AreYouIN in Facebook Apps

	fbsession := fb.NewSession(user.FbToken())
	fbFriends, err := fb.GetFriends(fbsession)
	if err != nil {
		return nil, errors.New(fb.GetErrorMessage(err))
	}

	// Match Facebook friends to AreYouIN users

	friends := make([]*UserAccount, 0, len(fbFriends))

	for _, fbFriend := range fbFriends {

		friend, err := self.userDAO.LoadByFB(fbFriend.Id)
		if err == api.ErrNotFound {
			// Skip: Facebook user has AreYouIN Facebook App but it's not registered (strangely)
			continue
		} else if err != nil {
			return nil, err
		}

		friends = append(friends, NewUserFromDTO(friend))
	}

	log.Printf("GetFacebookFriends: %v/%v friends found\n", len(friends), len(fbFriends))

	return friends, nil
}

// Adds Facebook friends using AreYouIN to user's friends list
// Returns the list of users that has been added by this operation
func (self *FriendManager) ImportFacebookFriends(user *UserAccount) ([]*UserAccount, error) {

	// Get areyouin accounts of Facebook friends
	facebookFriends, err := self.GetFacebookFriends(user)
	if err != nil {
		return nil, err
	}

	// Get existing friends
	storedFriends, err := self.friendDAO.LoadFriends(user.id, ALL_CONTACTS_GROUP)
	if err != nil {
		return nil, err
	}

	// Index friends
	friendsIndex := make(map[int64]*api.FriendDTO)
	for _, friend := range storedFriends {
		friendsIndex[friend.UserId] = friend
	}

	// Filter facebookFriends to get only new friends
	newFriends := make([]*UserAccount, 0, len(facebookFriends))

	for _, fbFriend := range facebookFriends {

		// Assume that if fbFriend isn't in storedFriends, then user wouldn't be either
		// in the fbFriend friends list
		if _, ok := friendsIndex[fbFriend.Id()]; !ok {

			if err := self.friendDAO.MakeFriends(user.AsFriend().AsDTO(), fbFriend.AsFriend().AsDTO()); err == nil {
				log.Printf("ImportFacebookFriends: %v and %v are now friends\n", user.Id(), fbFriend.Id())
				newFriends = append(newFriends, fbFriend)
			} else {
				// Log error but do not fail
				log.Printf("ImportFacebookFriends Error (userId=%v, friendId=%v): %v\n", user.Id(), fbFriend.Id(), err)
				continue
			}
		}
	}

	return newFriends, nil
}

// Since makeFriends() is bidirectional (adds the friend to user1 and user2). It can
// be assumed that if first user is friend of the second one, then second user must also
// have the first user in his/her friend list.
func (self *FriendManager) IsFriend(user1 int64, user2 int64) (bool, error) {
	if user1 == user2 {
		return true, nil
	}
	return self.friendDAO.ContainsFriend(user2, user1)
}

func (self *FriendManager) AreFriends(user1 int64, user2 int64) (bool, error) {

	ok, err := self.IsFriend(user1, user2)
	if err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}

	ok, err = self.IsFriend(user2, user1)
	if err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}

	return true, nil
}
