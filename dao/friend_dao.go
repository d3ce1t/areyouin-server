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
	session *gocql.Session
}

func NewFriendDAO(session *gocql.Session) core.FriendDAO {
	return &FriendDAO{session: session}
}

func (dao *FriendDAO) LoadFriends(user_id uint64, group_id uint32) ([]*core.Friend, error) {

	dao.checkSession()

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

func (dao *FriendDAO) LoadFriendsMap(user_id uint64) (map[uint64]*core.Friend, error) {

	dao.checkSession()

	stmt := `SELECT friend_id, friend_name, picture_digest FROM friends_by_user
						WHERE user_id = ? LIMIT ?`

	iter := dao.session.Query(stmt, int64(user_id), MAX_NUM_FRIENDS).Iter()

	friend_map := make(map[uint64]*core.Friend)

	var friend_id int64
	var friend_name string
	var picture_digest []byte

	for iter.Scan(&friend_id, &friend_name, &picture_digest) {
		friend_map[uint64(friend_id)] = &core.Friend{
			UserId:        uint64(friend_id),
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
func (dao *FriendDAO) IsFriend(user_id uint64, other_user_id uint64) (bool, error) {

	dao.checkSession()

	stmt := `SELECT user_id FROM friends_by_user
		WHERE user_id = ? AND friend_id = ?`

	err := dao.session.Query(stmt, int64(user_id), int64(other_user_id)).Scan(nil)

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
func (dao *FriendDAO) AreFriends(user_id uint64, other_user_id uint64) (bool, error) {

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

	dao.checkSession()

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	stmt := `INSERT INTO friends_by_user (user_id, friend_id, friend_name, picture_digest)
							VALUES (?, ?, ?, ?)`

	batch.Query(stmt, int64(user1.GetUserId()), int64(user2.GetUserId()), user2.GetName(), user2.GetPictureDigest())
	batch.Query(stmt, int64(user2.GetUserId()), int64(user1.GetUserId()), user1.GetName(), user1.GetPictureDigest())

	return dao.session.ExecuteBatch(batch)
}

func (dao *FriendDAO) SetPictureDigest(user_id uint64, friend_id uint64, digest []byte) error {
	dao.checkSession()
	stmt := `UPDATE friends_by_user SET picture_digest = ?
						WHERE user_id = ? AND friend_id = ?`
	return dao.session.Query(stmt, digest, int64(user_id), int64(friend_id)).Exec()
}

func (dao *FriendDAO) LoadGroups(user_id uint64) ([]*core.Group, error) {

	dao.checkSession()

	stmt := `SELECT group_id, group_name, group_size FROM groups_by_user
		WHERE user_id = ?`

	var group_id int32
	var group_name string
	var group_size int32
	groups := make([]*core.Group, 0, AVG_GROUPS_PER_USER)

	iter := dao.session.Query(stmt, int64(user_id)).Iter()

	for iter.Scan(&group_id, &group_name, &group_size) {
		groups = append(groups, &core.Group{
			Id:   uint32(group_id),
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

func (dao *FriendDAO) LoadGroupsAndMembers(user_id uint64) ([]*core.Group, error) {

	dao.checkSession()

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
func (dao *FriendDAO) AddGroup(user_id uint64, group *core.Group) error {

	dao.checkSession()

	stmt_add_group := `INSERT INTO groups_by_user (user_id, group_id, group_name, group_size)
		VALUES (?, ?, ?, ?)`

	stmt_add_member := `INSERT INTO friends_by_group (user_id, group_id, friend_id)
		VALUES (?, ?, ?)`

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	batch.Query(stmt_add_group, int64(user_id), int32(group.Id), group.Name, group.Size)

	for _, friend_id := range group.Members {
		batch.Query(stmt_add_member, int64(user_id), int32(group.Id), int64(friend_id))
	}

	return dao.session.ExecuteBatch(batch)
}

func (dao *FriendDAO) SetGroupName(user_id uint64, group_id uint32, name string) error {
	dao.checkSession()
	stmt_update := `UPDATE groups_by_user SET group_name = ?
		WHERE user_id = ? AND group_id = ?`
	return dao.session.Query(stmt_update, name, int64(user_id), int32(group_id)).Exec()
}

// Add one or more members into the same group
func (dao *FriendDAO) AddMembers(user_id uint64, group_id uint32, friend_ids ...uint64) error {

	dao.checkSession()

	if len(friend_ids) == 0 {
		return ErrInvalidArg
	}

	stmt := `INSERT INTO friends_by_group (user_id, group_id, friend_id)
		VALUES (?, ?, ?)`

	batch := dao.session.NewBatch(gocql.UnloggedBatch)

	for _, id := range friend_ids {
		batch.Query(stmt, int64(user_id), int32(group_id), id)
	}

	err := dao.session.ExecuteBatch(batch)
	if err != nil {
		return err
	}

	return dao.updateGroupSize(user_id, group_id)
}

// Delete one or more members from the same group
func (dao *FriendDAO) DeleteMembers(user_id uint64, group_id uint32, friend_ids ...uint64) error {

	dao.checkSession()

	if len(friend_ids) == 0 {
		return ErrInvalidArg
	}

	stmt := `DELETE FROM friends_by_group WHERE user_id = ? AND group_id = ? AND friend_id = ?`

	batch := dao.session.NewBatch(gocql.UnloggedBatch)

	for _, id := range friend_ids {
		batch.Query(stmt, int64(user_id), int32(group_id), int64(id))
	}

	err := dao.session.ExecuteBatch(batch)
	if err != nil {
		return err
	}

	return dao.updateGroupSize(user_id, group_id)
}

func (dao *FriendDAO) DeleteGroup(user_id uint64, group_id uint32) error {

	dao.checkSession()

	stmt_empty_group := `DELETE FROM friends_by_group WHERE user_id = ? AND group_id = ?`
	stmt_delete_group := `DELETE FROM groups_by_user WHERE user_id = ? AND group_id = ?`

	batch := dao.session.NewBatch(gocql.LoggedBatch)
	batch.Query(stmt_empty_group, int64(user_id), int32(group_id))
	batch.Query(stmt_delete_group, int64(user_id), int32(group_id))

	return dao.session.ExecuteBatch(batch)
}

func (dao *FriendDAO) checkSession() {
	if dao.session == nil {
		panic(ErrNoSession)
	}
}
