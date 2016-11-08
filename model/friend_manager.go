package model

import (
	"errors"
	"log"
	"peeple/areyouin/api"
	"peeple/areyouin/cqldao"
	fb "peeple/areyouin/facebook"

	observer "github.com/imkira/go-observer"
)

type FriendManager struct {
	dbsession        api.DbSession
	parent           *AyiModel
	userDAO          api.UserDAO
	friendDAO        api.FriendDAO
	friendRequestDAO api.FriendRequestDAO
	friendSignal     observer.Property
}

func newFriendManager(parent *AyiModel, session api.DbSession) *FriendManager {
	return &FriendManager{
		parent:           parent,
		dbsession:        session,
		userDAO:          cqldao.NewUserDAO(session),
		friendDAO:        cqldao.NewFriendDAO(session),
		friendRequestDAO: cqldao.NewFriendRequestDAO(session),
		friendSignal:     observer.NewProperty(nil),
	}
}

func (m *FriendManager) Observe() observer.Stream {
	return m.friendSignal.Observe()
}

func (self *FriendManager) GetAllFriends(userID int64) ([]*Friend, error) {

	friendsDTO, err := self.friendDAO.LoadFriends(userID, allContactsGroup)
	if err != nil {
		return nil, err
	}

	friends := make([]*Friend, 0, len(friendsDTO))
	for _, f := range friendsDTO {
		friends = append(friends, newFriendFromDTO(f))
	}

	return friends, nil
}

func (m *FriendManager) GetAllGroups(userID int64) ([]*Group, error) {

	groupsDTO, err := m.friendDAO.LoadGroupsWithMembers(userID)
	if err != nil {
		return nil, err
	}

	groups := newGroupListFromDTO(groupsDTO)

	return groups, nil
}

// GetFacebookFriends gets AreYouIN users that are friends of given user in Facebook
func (m *FriendManager) GetFacebookFriends(user *UserAccount) ([]*UserAccount, error) {

	if user.FbId() == "" || user.FbToken() == "" {
		return nil, ErrAccountNotLinkedToFacebook
	}

	// Check access, i.e. access token is for fbId
	fbsession := fb.NewSession(user.FbToken())
	if _, err := fb.CheckAccess(user.FbId(), fbsession); err != nil {
		return nil, err
	}

	// Load Facebook friends that have AreYouIN in Facebook Apps

	fbFriends, err := fb.GetFriends(fbsession)
	if err != nil {
		return nil, errors.New(fb.GetErrorMessage(err))
	}

	// Match Facebook friends to AreYouIN users

	users := make([]*UserAccount, 0, len(fbFriends))

	for _, fbFriend := range fbFriends {

		friend, err := m.userDAO.LoadByFB(fbFriend.Id)
		if err == api.ErrNotFound {
			// Skip: Facebook user has AreYouIN Facebook App but it's not registered (strangely)
			continue
		} else if err != nil {
			return nil, err
		}

		users = append(users, newUserFromDTO(friend))
	}

	log.Printf("GetFacebookFriends: %v/%v friends found\n", len(users), len(fbFriends))

	return users, nil
}

// GetNewFacebookFriends gets AreYouIN users that are friends in Facebook but not in AreYouIN
func (m *FriendManager) GetNewFacebookFriends(user *UserAccount) ([]*UserAccount, error) {

	// Get areyouin accounts of Facebook friends
	facebookFriends, err := m.GetFacebookFriends(user)
	if err != nil {
		return nil, err
	}

	// Get existing friends
	storedFriends, err := m.friendDAO.LoadFriends(user.id, allContactsGroup)
	if err != nil {
		return nil, err
	}

	// Index friends
	friendsIndex := make(map[int64]*api.FriendDTO)
	for _, friend := range storedFriends {
		friendsIndex[friend.UserId] = friend
	}

	// Filter facebookFriends to get only new friends
	newFBFriends := make([]*UserAccount, 0, len(facebookFriends))

	for _, fbFriend := range facebookFriends {
		// Assume that if fbFriend isn't in storedFriends, then user wouldn't be either
		// in the fbFriend friends list
		if _, ok := friendsIndex[fbFriend.Id()]; !ok {
			newFBFriends = append(newFBFriends, fbFriend)
		}
	}

	return newFBFriends, nil
}

// ImportFacebookFriends adds to user's list those AreYouIN users that are friends in Facebook but not in AreYouIN
// Returns the list of users that has been added by this operation
func (m *FriendManager) ImportFacebookFriends(user *UserAccount, initialImport bool) ([]*UserAccount, error) {

	newFacebookFriends, err := m.GetNewFacebookFriends(user)
	if err != nil {
		return nil, err
	}

	addedFriends := make([]*UserAccount, 0, len(newFacebookFriends))

	for _, fbFriend := range newFacebookFriends {

		if err := m.friendDAO.MakeFriends(user.AsFriend().AsDTO(), fbFriend.AsFriend().AsDTO()); err == nil {
			log.Printf("ImportFacebookFriends: %v and %v are now friends\n", user.Id(), fbFriend.Id())
			addedFriends = append(addedFriends, fbFriend)
		} else {
			// Log error but do not fail
			log.Printf("ImportFacebookFriends Error (userId=%v, friendId=%v): %v\n", user.Id(), fbFriend.Id(), err)
			continue
		}
	}

	if len(addedFriends) > 0 {
		signal := &Signal{
			Type: SignalNewFriendsImported,
			Data: map[string]interface{}{
				"User":          user,
				"NewFriends":    addedFriends,
				"InitialImport": initialImport,
			},
		}

		m.friendSignal.Update(signal)
	}

	return addedFriends, nil
}

func (m *FriendManager) MakeFriends(user1 *UserAccount, user2 *UserAccount) error {

	err := m.friendDAO.MakeFriends(user1.AsFriend().AsDTO(), user2.AsFriend().AsDTO())
	if err != nil {
		return err
	}

	/*signal := &Signal{
		Type: SIGNAL_NEW_FRIENDS,
		Data: map[string]interface{}{
			"user1": user1.Id(),
			"user2": user2.Id(),
		},
	}

	m.friendSignal.Update(signal)*/

	return nil
}

// Since makeFriends() is bidirectional (adds the friend to user1 and user2). It can
// be assumed that if first user is friend of the second one, then second user must also
// have the first user in his/her friend list.
func (m *FriendManager) IsFriend(user1 int64, user2 int64) (bool, error) {
	if user1 == user2 {
		return true, nil
	}
	return m.friendDAO.ContainsFriend(user2, user1)
}

func (m *FriendManager) AreFriends(user1 int64, user2 int64) (bool, error) {

	ok, err := m.IsFriend(user1, user2)
	if err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}

	ok, err = m.IsFriend(user2, user1)
	if err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}

	return true, nil
}

/*func (m *FriendManager) GetFriendRequest(fromUser int64, toUser int64) (*FriendRequest, error) {

	friendRequestDTO, err := m.friendRequestDAO.Load(fromUser, toUser)
	if err != nil {
		return nil, err
	}

	return newFriendRequestFromDTO(friendRequestDTO), nil
}*/

func (m *FriendManager) GetAllFriendRequests(toUser int64) ([]*FriendRequest, error) {

	friendRequestsDTO, err := m.friendRequestDAO.LoadAll(toUser)
	if err != nil {
		return nil, err
	}

	return newFriendRequestListFromDTO(friendRequestsDTO), nil
}

func (m *FriendManager) CreateFriendRequest(fromUser *UserAccount, toUser *UserAccount) (*FriendRequest, error) {

	// Not friends
	if areFriends, err := m.parent.Friends.AreFriends(fromUser.Id(), toUser.Id()); err != nil {
		return nil, err
	} else if areFriends {
		return nil, ErrAlreadyFriends
	}

	// Check friend request exist
	if exist, err := m.friendRequestDAO.Exist(fromUser.Id(), toUser.Id()); err != nil {
		return nil, err
	} else if exist {
		return nil, ErrFriendRequestAlreadyExist
	}

	// Add friend request
	friendRequest := NewFriendRequest(toUser.Id(), fromUser.Id(), fromUser.name, fromUser.email)
	if err := m.friendRequestDAO.Insert(friendRequest.AsDTO()); err != nil {
		return nil, err
	}

	signal := &Signal{
		Type: SignalNewFriendRequest,
		Data: map[string]interface{}{
			"FromUser":      fromUser,
			"ToUser":        toUser,
			"FriendRequest": friendRequest,
		},
	}

	m.friendSignal.Update(signal)

	return friendRequest, nil
}

func (m *FriendManager) ConfirmFriendRequest(fromUser *UserAccount, toUser *UserAccount, accept bool) error {

	friendRequestDTO, err := m.friendRequestDAO.Load(fromUser.Id(), toUser.Id())
	if err != nil {
		return err
	}

	if accept {

		// Make both friends
		err := m.friendDAO.MakeFriends(toUser.AsFriend().AsDTO(), fromUser.AsFriend().AsDTO())
		if err != nil {
			return err
		}

		err = m.friendRequestDAO.Delete(friendRequestDTO)
		if err != nil {
			return err
		}

		signal := &Signal{
			Type: SignalFriendRequestAccepted,
			Data: map[string]interface{}{
				"FromUser": fromUser,
				"ToUser":   toUser,
			},
		}

		m.friendSignal.Update(signal)

	} else {

		err = m.friendRequestDAO.Delete(friendRequestDTO)
		if err != nil {
			return err
		}

		signal := &Signal{
			Type: SignalFriendRequestCancelled,
			Data: map[string]interface{}{
				"FromUser": fromUser,
				"ToUser":   toUser,
			},
		}

		m.friendSignal.Update(signal)
	}

	return nil
}

/*func (m *FriendManager) ExistFriendRequest(fromUser int64, toUser int64) (bool, error) {
	return m.friendRequestDAO.Exist(fromUser, toUser)
}*/

// Adds friends groups to user 'userId'. If a group already exists it is updated.
func (m *FriendManager) AddGroups(userID int64, groups []*Group) error {

	// Load server groups
	// TODO: All groups are always loaded. However, a subset could be loaded when sync
	// behaviour is not TRUNCATE.
	serverGroupsDTO, err := m.friendDAO.LoadGroupsWithMembers(userID)
	if err != nil {
		return err
	}

	// Slice of GroupDTO to Slice of Group
	serverGroups := newGroupListFromDTO(serverGroupsDTO)

	// Add
	m.addFriendGroups(userID, serverGroups, groups)

	return nil
}

// Remove group
func (m *FriendManager) DeleteGroup(userID int64, groupID int32) error {
	return m.friendDAO.DeleteGroup(userID, groupID)
}

// Rename group
func (m *FriendManager) RenameGroup(userID int64, groupID int32, newName string) error {
	return m.friendDAO.SetGroupName(userID, groupID, newName)
}

// Add groups. If a client group already exists in server it is updated.
func (m *FriendManager) addFriendGroups(userID int64, serverGroups []*Group,
	clientGroups []*Group) error {

	// Index groups
	clientGroupsIndex := make(map[int32]*Group)
	for _, group := range clientGroups {
		clientGroupsIndex[group.id] = group
	}

	// Loop through server groups in order to know what
	// to do: update/replace or remove group from server
	for _, group := range serverGroups {

		if clientGroup, ok := clientGroupsIndex[group.id]; ok {

			// Group exists. Update.

			if group.name != clientGroup.name {
				err := m.friendDAO.SetGroupName(userID, group.id, clientGroup.name)
				if err != nil {
					return err
				}
			}

			m.syncGroupMembers(userID, group.id, group.members, clientGroup.members)

			// Delete also from copy because it has been processed
			delete(clientGroupsIndex, group.id)

		}
	}

	// clientGroupsIndex contains only new groups.

	for _, group := range clientGroupsIndex {

		// Filter groups to remove non-friends.
		err := m.removeNonFriends(group, userID)
		if err != nil {
			return err
		}

		// Add group
		err = m.friendDAO.InsertGroup(userID, group.AsDTO())
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *FriendManager) removeNonFriends(group *Group, userID int64) error {

	newMembers := make([]int64, 0, group.size)

	for _, friendID := range group.members {

		isFriend, err := m.IsFriend(friendID, userID)
		if err != nil {
			return err
		}

		if isFriend {
			newMembers = append(newMembers, friendID)
		}
	}

	group.members = newMembers
	group.size = len(newMembers)

	return nil
}

func (m *FriendManager) syncGroupMembers(userID int64, groupID int32, serverMembers []int64, clientMembers []int64) error {

	// Index client members
	index := make(map[int64]bool)
	for _, id := range clientMembers {
		index[id] = true
	}

	// Determine which members (from serverMembers) have to be removed,
	// i.e. server members that no longer exists in client will be removed
	removeIds := make([]int64, 0, len(serverMembers)/2)

	for _, serverMemberId := range serverMembers {
		if _, ok := index[serverMemberId]; !ok {
			// mark for remove
			removeIds = append(removeIds, serverMemberId)
		} else {
			// if exist, delete from index so index will have only new members
			delete(index, serverMemberId)
		}
	}

	// Determine which new members (from clientMembers) have to be added,
	// i.e. Only those that are actually friends
	idsToAdd := make([]int64, 0, len(clientMembers))

	for clientMemberID := range index {

		isFriend, err := m.IsFriend(clientMemberID, userID)
		if err != nil {
			return err
		}

		if isFriend {
			idsToAdd = append(idsToAdd, clientMemberID)
		}
	}

	// Proceed database I/O
	if len(removeIds) > 0 {
		err := m.friendDAO.DeleteMembers(userID, groupID, removeIds...)
		if err != nil {
			return err
		}
	}

	if len(idsToAdd) > 0 {
		err := m.friendDAO.AddMembers(userID, groupID, idsToAdd...)
		if err != nil {
			return err
		}
	}

	return nil
}
