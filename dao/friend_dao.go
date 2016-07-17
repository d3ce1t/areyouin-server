package dao

import (
	"log"
	core "peeple/areyouin/common"
	"github.com/gocql/gocql"
)

const (
	MAX_NUM_FRIENDS     = 1000
	AVG_GROUPS_PER_USER = 10
)

type FriendDAO struct {
	session *GocqlSession
}

//
// Friends
//

func (dao *FriendDAO) LoadFriends(user_id int64, group_id int32) ([]*core.Friend, error) {

	checkSession(dao.session)

	if group_id == 0 {
		return dao.getAllFriends(user_id)
	} else {
		friends_id, err := dao.getFriendsIdInGroup(user_id, group_id)
		if err != nil {
			return nil, err
		}
		return dao.getFriends(user_id, friends_id...)
	}
}

func (dao *FriendDAO) LoadFriendsMap(user_id int64) (map[int64]*core.Friend, error) {

	checkSession(dao.session)

	stmt := `SELECT friend_id, friend_name, picture_digest FROM friends_by_user
						WHERE user_id = ? LIMIT ?`

	iter := dao.session.Query(stmt, user_id, MAX_NUM_FRIENDS).Iter()

	friend_map := make(map[int64]*core.Friend)

	var friend_id int64
	var friend_name string
	var picture_digest []byte

	for iter.Scan(&friend_id, &friend_name, &picture_digest) {
		friend_map[friend_id] = &core.Friend{
			UserId:        friend_id,
			Name:          friend_name,
			PictureDigest: picture_digest,
		}
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadFriendsMap:", err)
		return nil, err
	}

	return friend_map, nil
}

// Since makeFriends() is bidirectional (adds the friend to user1 and user2). It can
// be assumed that if first user is friend of the second one, then second user must also
// have the first user in his/her friend list.
func (dao *FriendDAO) IsFriend(user_id int64, other_user_id int64) (bool, error) {

	checkSession(dao.session)

	stmt := `SELECT user_id FROM friends_by_user
		WHERE user_id = ? AND friend_id = ?`

	err := dao.session.Query(stmt, user_id, other_user_id).Scan(nil)

	if err != nil { // HACK: 0 group contains ALL_CONTACTS
		if err == ErrNotFound {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

// Check if two users are friends. In contrast to IsFriend. This function perform checking
// in two ways.
func (dao *FriendDAO) AreFriends(user_id int64, other_user_id int64) (bool, error) {

	var one_way bool
	var err error
	var two_way bool

	if one_way, err = dao.IsFriend(user_id, other_user_id); one_way {
		if two_way, err = dao.IsFriend(other_user_id, user_id); two_way {
			return true, nil
		}
	}

	return false, err
}

func (dao *FriendDAO) MakeFriends(user1 core.UserFriend, user2 core.UserFriend) error {

	checkSession(dao.session)

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	stmt := `INSERT INTO friends_by_user (user_id, friend_id, friend_name, picture_digest)
							VALUES (?, ?, ?, ?)`

	batch.Query(stmt, user1.GetUserId(), user2.GetUserId(), user2.GetName(), user2.GetPictureDigest())
	batch.Query(stmt, user2.GetUserId(), user1.GetUserId(), user1.GetName(), user1.GetPictureDigest())

	return dao.session.ExecuteBatch(batch)
}

func (dao *FriendDAO) SetPictureDigest(user_id int64, friend_id int64, digest []byte) error {

	checkSession(dao.session)

	stmt := `UPDATE friends_by_user SET picture_digest = ?
						WHERE user_id = ? AND friend_id = ?`

	return dao.session.Query(stmt, digest, user_id, friend_id).Exec()
}

//
// Groups
//

func (dao *FriendDAO) LoadGroups(user_id int64) ([]*core.Group, error) {

	checkSession(dao.session)

	stmt := `SELECT group_id, group_name, group_size FROM groups_by_user
		WHERE user_id = ?`

	var group_id int32
	var group_name string
	var group_size int32
	groups := make([]*core.Group, 0, AVG_GROUPS_PER_USER)

	iter := dao.session.Query(stmt, user_id).Iter()

	for iter.Scan(&group_id, &group_name, &group_size) {
		groups = append(groups, &core.Group{
			Id:   group_id,
			Name: group_name,
			Size: group_size,
		})
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadGroups:", err)
		return nil, err
	}

	return groups, nil
}

func (dao *FriendDAO) LoadGroupsAndMembers(user_id int64) ([]*core.Group, error) {

	checkSession(dao.session)

	groups, err := dao.LoadGroups(user_id)
	if err != nil {
		return nil, err
	}

	if len(groups) > 0 {
		if err = dao.loadMembersIntoGroups(user_id, groups); err != nil {
			return nil, err
		}
	}

	return groups, nil
}

// Add a new group with members. If group contains some members, they must be all
// of the members in the group.
func (dao *FriendDAO) AddGroup(user_id int64, group *core.Group) error {

	checkSession(dao.session)

	stmt_add_group := `INSERT INTO groups_by_user (user_id, group_id, group_name, group_size)
		VALUES (?, ?, ?, ?)`

	stmt_add_member := `INSERT INTO friends_by_group (user_id, group_id, friend_id)
		VALUES (?, ?, ?)`

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(stmt_add_group, user_id, group.Id, group.Name, group.Size)

	for _, friend_id := range group.Members {
		batch.Query(stmt_add_member, user_id, group.Id, friend_id)
	}

	return dao.session.ExecuteBatch(batch)
}

func (dao *FriendDAO) SetGroupName(user_id int64, group_id int32, name string) error {

	checkSession(dao.session)

	stmt_update := `UPDATE groups_by_user SET group_name = ?
		WHERE user_id = ? AND group_id = ?`

	return dao.session.Query(stmt_update, name, user_id, group_id).Exec()
}

// Add one or more members into the same group
func (dao *FriendDAO) AddMembers(user_id int64, group_id int32, friend_ids ...int64) error {

	checkSession(dao.session)

	if len(friend_ids) == 0 {
		return ErrInvalidArg
	}

	stmt := `INSERT INTO friends_by_group (user_id, group_id, friend_id)
		VALUES (?, ?, ?)`

	batch := dao.session.NewBatch(gocql.UnloggedBatch)

	for _, id := range friend_ids {
		batch.Query(stmt, user_id, group_id, id)
	}

	err := dao.session.ExecuteBatch(batch)
	if err != nil {
		return err
	}

	return dao.updateGroupSize(user_id, group_id)
}

// Delete one or more members from the same group
func (dao *FriendDAO) DeleteMembers(user_id int64, group_id int32, friend_ids ...int64) error {

	checkSession(dao.session)

	if len(friend_ids) == 0 {
		return ErrInvalidArg
	}

	stmt := `DELETE FROM friends_by_group WHERE user_id = ? AND group_id = ? AND friend_id = ?`

	batch := dao.session.NewBatch(gocql.UnloggedBatch)

	for _, id := range friend_ids {
		batch.Query(stmt, user_id, group_id, id)
	}

	err := dao.session.ExecuteBatch(batch)
	if err != nil {
		return err
	}

	return dao.updateGroupSize(user_id, group_id)
}

func (dao *FriendDAO) DeleteGroup(user_id int64, group_id int32) error {

	checkSession(dao.session)

	stmt_empty_group := `DELETE FROM friends_by_group WHERE user_id = ? AND group_id = ?`
	stmt_delete_group := `DELETE FROM groups_by_user WHERE user_id = ? AND group_id = ?`

	batch := dao.session.NewBatch(gocql.LoggedBatch)
	batch.Query(stmt_empty_group, user_id, group_id)
	batch.Query(stmt_delete_group, user_id, group_id)

	return dao.session.ExecuteBatch(batch)
}

//
// Friend requests
//

func (dao *FriendDAO) LoadFriendRequest(user_id int64, friend_id int64) (*core.FriendRequest, error) {

	checkSession(dao.session)

	if user_id == 0 || friend_id == 0 {
		return nil, ErrNotFound
	}

	// Load created date from friend_requests_sent
	stmt_requests_sent := `SELECT created_date FROM friend_requests_sent
	 	WHERE user_id = ? AND friend_id = ? LIMIT 1`

	var created_date int64

	if err := dao.session.Query(stmt_requests_sent, friend_id, user_id).Scan(&created_date); err != nil {
		return nil, err
	}

	// Now load friend request in friend_requests_received
	stmt_requests_received := `SELECT name, email FROM friend_requests_received
		WHERE user_id = ? AND created_date = ? AND friend_id = ?`

	var name string
	var email string

	q := dao.session.Query(stmt_requests_received, user_id, created_date, friend_id)
	if err := q.Scan(&name, &email); err != nil {
		return nil, err
	}

	// All data has been read
	friendRequest := &core.FriendRequest {
		FriendId: friend_id,
		Name: name,
		Email: email,
		CreatedDate: created_date,
	}

	return friendRequest, nil
}


func (dao *FriendDAO) LoadFriendRequests(user_id int64) ([]*core.FriendRequest, error) {

	if user_id == 0 {
		return nil, ErrNotFound
	}

	checkSession(dao.session)

	stmt := `SELECT created_date, friend_id, name, email FROM friend_requests_received
		WHERE user_id = ?`

	iter := dao.session.Query(stmt, user_id).Iter()

	requests := make([]*core.FriendRequest, 0, 10)

	var created_date int64
	var friend_id int64
	var name string
	var email string

	for iter.Scan(&created_date, &friend_id, &name, &email) {
		requests = append(requests, &core.FriendRequest{
			FriendId: friend_id,
			Name: name,
			Email: email,
			CreatedDate: created_date,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return requests, nil
}

// Check if exists a friend request from user_id to friend_id. In other words, if user_id
// has already sent (or not) a friend request to friend_id.
func (dao *FriendDAO) ExistFriendRequest(user_id int64, friend_id int64) (bool, error) {

	checkSession(dao.session)

	if user_id == 0 || friend_id == 0 {
		return false, ErrInvalidArg
	}

	stmt := `SELECT created_date FROM friend_requests_sent
	 	WHERE user_id = ? AND friend_id = ? LIMIT 1`

	var created_date int64

	if err := dao.session.Query(stmt, user_id, friend_id).Scan(&created_date); err != nil {

		if err == ErrNotFound {
			return false, nil
		} else {
			return false, err
		}

	}

	return true, nil
}

// Insert a friend request into database. This request means that some user (friend_id)
// wants to be friend of user_id
func (dao *FriendDAO) InsertFriendRequest(user_id int64, friend_id int64, name string, email string, created_date int64) error {

	checkSession(dao.session)

	if user_id == 0 || friend_id == 0 {
		return ErrInvalidArg
	}

	stmt_insert_sent := `INSERT INTO friend_requests_sent (user_id, friend_id, created_date)
		VALUES (?, ?, ?)`

	stmt_insert_received := `INSERT INTO friend_requests_received (user_id, created_date, friend_id, name, email)
		VALUES (?, ?, ?, ?, ?)`

	batch := dao.session.NewBatch(gocql.LoggedBatch)
	batch.Query(stmt_insert_sent, friend_id, user_id, created_date)
	batch.Query(stmt_insert_received, user_id, created_date, friend_id, name, email)

	return dao.session.ExecuteBatch(batch)
}

func (dao *FriendDAO) DeleteFriendRequest(user_id int64, friend_id int64, created_date int64) error {

	checkSession(dao.session)

	if user_id == 0 || friend_id == 0 {
		return ErrInvalidArg
	}

	stmt_delete_sent := `DELETE FROM friend_requests_sent WHERE user_id = ? AND friend_id = ?`
	stmt_delete_received := `DELETE FROM friend_requests_received
		WHERE user_id = ? AND created_date = ? AND friend_id = ?`

	batch := dao.session.NewBatch(gocql.LoggedBatch)
	batch.Query(stmt_delete_sent, friend_id, user_id)
	batch.Query(stmt_delete_received, user_id, created_date, friend_id)

	return dao.session.ExecuteBatch(batch)
}
