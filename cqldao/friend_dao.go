package cqldao

import (
	"log"
	"peeple/areyouin/api"

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

func (d *FriendDAO) LoadFriends(user_id int64, group_id int32) ([]*api.FriendDTO, error) {

	checkSession(d.session)

	if group_id == 0 {
		return d.getAllFriends(user_id)
	} else {
		friends_id, err := d.getFriendsIdInGroup(user_id, group_id)
		if err != nil {
			return nil, err
		}
		return d.getFriends(user_id, friends_id...)
	}
}

func (dao *FriendDAO) ContainsFriend(user_id int64, other_user_id int64) (bool, error) {

	checkSession(dao.session)

	stmt := `SELECT user_id FROM friends_by_user
		WHERE user_id = ? AND friend_id = ?`

	err := dao.session.Query(stmt, user_id, other_user_id).Scan(nil)

	if err == gocql.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, convErr(err)
	}

	return true, nil
}

func (dao *FriendDAO) MakeFriends(user1 *api.FriendDTO, user2 *api.FriendDTO) error {

	checkSession(dao.session)

	batch := dao.session.NewBatch(gocql.LoggedBatch)

	stmt := `INSERT INTO friends_by_user (user_id, friend_id, friend_name, picture_digest)
							VALUES (?, ?, ?, ?)`

	batch.Query(stmt, user1.UserId, user2.UserId, user2.Name, user2.PictureDigest)
	batch.Query(stmt, user2.UserId, user1.UserId, user1.Name, user1.PictureDigest)

	return convErr(dao.session.ExecuteBatch(batch))
}

func (dao *FriendDAO) SetFriendPictureDigest(user_id int64, friend_id int64, digest []byte) error {

	checkSession(dao.session)

	stmt := `UPDATE friends_by_user SET picture_digest = ?
						WHERE user_id = ? AND friend_id = ?`

	err := dao.session.Query(stmt, digest, user_id, friend_id).Exec()

	return convErr(err)
}

//
// Groups
//

func (d *FriendDAO) LoadGroups(user_id int64) ([]*api.GroupDTO, error) {

	checkSession(d.session)

	stmt := `SELECT group_id, group_name, group_size FROM groups_by_user
		WHERE user_id = ?`

	var group_id int32
	var group_name string
	var group_size int32
	groups := make([]*api.GroupDTO, 0, AVG_GROUPS_PER_USER)

	iter := d.session.Query(stmt, user_id).Iter()

	for iter.Scan(&group_id, &group_name, &group_size) {
		groups = append(groups, &api.GroupDTO{
			Id:   group_id,
			Name: group_name,
			Size: group_size,
		})
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadGroups:", err)
		return nil, convErr(err)
	}

	return groups, nil
}

func (dao *FriendDAO) LoadGroupsWithMembers(user_id int64) ([]*api.GroupDTO, error) {

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

// Add a new group with members. If a group contains members, they must be all
// of the members in the group.
func (dao *FriendDAO) InsertGroup(user_id int64, group *api.GroupDTO) error {

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

	return convErr(dao.session.ExecuteBatch(batch))
}

func (dao *FriendDAO) SetGroupName(user_id int64, group_id int32, name string) error {

	checkSession(dao.session)

	stmt_update := `UPDATE groups_by_user SET group_name = ?
		WHERE user_id = ? AND group_id = ?`

	err := dao.session.Query(stmt_update, name, user_id, group_id).Exec()
	return convErr(err)
}

// Add one or more members into the same group
func (d *FriendDAO) AddMembers(user_id int64, group_id int32, friend_ids ...int64) error {

	checkSession(d.session)

	if len(friend_ids) == 0 {
		return api.ErrInvalidArg
	}

	stmt := `INSERT INTO friends_by_group (user_id, group_id, friend_id)
		VALUES (?, ?, ?)`

	batch := d.session.NewBatch(gocql.UnloggedBatch)

	for _, id := range friend_ids {
		batch.Query(stmt, user_id, group_id, id)
	}

	if err := d.session.ExecuteBatch(batch); err != nil {
		return convErr(err)
	}

	return d.updateGroupSize(user_id, group_id)
}

// Delete one or more members from the same group
func (d *FriendDAO) DeleteMembers(user_id int64, group_id int32, friend_ids ...int64) error {

	checkSession(d.session)

	if len(friend_ids) == 0 {
		return api.ErrInvalidArg
	}

	stmt := `DELETE FROM friends_by_group WHERE user_id = ? AND group_id = ? AND friend_id = ?`

	batch := d.session.NewBatch(gocql.UnloggedBatch)

	for _, id := range friend_ids {
		batch.Query(stmt, user_id, group_id, id)
	}

	if err := d.session.ExecuteBatch(batch); err != nil {
		return convErr(err)
	}

	return d.updateGroupSize(user_id, group_id)
}

func (dao *FriendDAO) DeleteGroup(user_id int64, group_id int32) error {

	checkSession(dao.session)

	stmt_empty_group := `DELETE FROM friends_by_group WHERE user_id = ? AND group_id = ?`
	stmt_delete_group := `DELETE FROM groups_by_user WHERE user_id = ? AND group_id = ?`

	batch := dao.session.NewBatch(gocql.LoggedBatch)
	batch.Query(stmt_empty_group, user_id, group_id)
	batch.Query(stmt_delete_group, user_id, group_id)

	return convErr(dao.session.ExecuteBatch(batch))
}
