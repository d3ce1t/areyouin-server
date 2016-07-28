package model

import (
	"errors"
	"log"
	"peeple/areyouin/api"
	"peeple/areyouin/cqldao"
	fb "peeple/areyouin/facebook"
)

type FriendManager struct {
	dbsession        api.DbSession
	parent           *AyiModel
	userDAO          api.UserDAO
	friendDAO        api.FriendDAO
	friendRequestDAO api.FriendRequestDAO
}

func newFriendManager(parent *AyiModel, session api.DbSession) *FriendManager {
	return &FriendManager{
		parent:           parent,
		dbsession:        session,
		userDAO:          cqldao.NewUserDAO(session),
		friendDAO:        cqldao.NewFriendDAO(session),
		friendRequestDAO: cqldao.NewFriendRequestDAO(session),
	}
}

func (self *FriendManager) GetAllFriends(userID int64) ([]*Friend, error) {

	friendsDTO, err := self.friendDAO.LoadFriends(userID, ALL_CONTACTS_GROUP)
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

		friends = append(friends, newUserFromDTO(friend))
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

func (m *FriendManager) MakeFriends(user1 *UserAccount, user2 *UserAccount) error {
	return m.friendDAO.MakeFriends(user1.AsFriend().AsDTO(), user2.AsFriend().AsDTO())
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

func (m *FriendManager) SendFriendRequest(fromUser *UserAccount, toUser *UserAccount) (*FriendRequest, error) {

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
	err := m.friendRequestDAO.Insert(friendRequest.AsDTO())

	return nil, err
}

func (m *FriendManager) ConfirmFriendRequest(fromUser *UserAccount, toUser *UserAccount, accept bool) error {

	friendRequestDTO, err := m.friendRequestDAO.Load(fromUser.Id(), toUser.Id())
	if err != nil {
		return err
	}

	if accept {

		// Make both friends
		err = m.parent.Friends.MakeFriends(toUser, fromUser)
		if err != nil {
			return err
		}
	}

	// Accepted or cancelled. Remove it.

	err = m.friendRequestDAO.Delete(friendRequestDTO)
	if err != nil {
		return err
	}

	return nil
}

/*func (m *FriendManager) ExistFriendRequest(fromUser int64, toUser int64) (bool, error) {
	return m.friendRequestDAO.Exist(fromUser, toUser)
}*/

// Adds friends groups to user 'userId'
func (m *FriendManager) SyncGroups(userID int64, groups []*Group) error {

	// Load server groups
	// TODO: All groups are always loaded. However, a subset could be loaded when sync
	// behaviour is not TRUNCATE.
	serverGroupsDTO, err := m.friendDAO.LoadGroupsWithMembers(userID)
	if err != nil {
		return err
	}

	// Slice of GroupDTO to Slice of Group
	serverGroups := newGroupListFromDTO(serverGroupsDTO)

	// Sync
	m.syncFriendGroups(userID, serverGroups, groups)

	return nil
}

// Sync server-side groups with client-side groups. If groups provided by the client
// contains all of the groups, then a full sync is required, i.e. server-side groups
// that does not exist client side are removed. Otherwise, if provided groups are only
// a subset, a merge of client and server data is performed. Conversely to full sync,
// merging process does not remove existing groups from the server but add new groups
// and modify existing ones. Regarding full sync, it is assumed that clientGroups
// contains all of the groups in client. Hence, if a group doesn't exist in client,
// it will be removed from server. Like a regular sync, new groups in client will be
// added to server. In whatever case, if a group already exists server-side, it will
// be updated with members from client-side group, removing those members that does not
// exist in client's group (client is master). In other words, groups at server will be
// equal to groups at client at the end of the synchronisation process.
func (m *FriendManager) syncFriendGroups(userID int64, serverGroups []*Group,
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

			// Group exists.

			if clientGroup.size == -1 && len(clientGroup.members) == 0 {

				// Special case

				if clientGroup.name == "" {

					// Group is marked for removal. So remove it from server

					err := m.friendDAO.DeleteGroup(userID, group.id)
					if err != nil {
						return err
					}

				} else if group.name != clientGroup.name {

					// Only Rename group

					err := m.friendDAO.SetGroupName(userID, group.id, clientGroup.name)
					if err != nil {
						return err
					}
				}

			} else {

				// Update case

				if group.name != clientGroup.name {
					err := m.friendDAO.SetGroupName(userID, group.id, clientGroup.name)
					if err != nil {
						return err
					}
				}

				m.syncGroupMembers(userID, group.id, group.members, clientGroup.members)
			}

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
