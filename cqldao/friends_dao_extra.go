package cqldao

import (
	"log"

	"github.com/d3ce1t/areyouin-server/api"
	"github.com/d3ce1t/areyouin-server/utils"
)

func (dao *FriendDAO) getFriendsIdInGroup(user_id int64, group_id int32) ([]int64, error) {

	checkSession(dao.session)

	stmt := `SELECT friend_id FROM friends_by_group
		WHERE user_id = ? AND group_id = ?`

	ids_slice := make([]int64, 0, 20)
	var friend_id int64

	iter := dao.session.Query(stmt, user_id, group_id).Iter()

	for iter.Scan(&friend_id) {
		ids_slice = append(ids_slice, friend_id)
	}

	if err := iter.Close(); err != nil {
		log.Println("getFriendsIdInGroup:", err)
		return nil, convErr(err)
	}

	return ids_slice, nil
}

func (d *FriendDAO) getAllFriends(user_id int64) ([]*api.FriendDTO, error) {

	checkSession(d.session)

	stmt := `SELECT friend_id, friend_name, picture_digest FROM friends_by_user
						WHERE user_id = ? LIMIT ?`

	iter := d.session.Query(stmt, user_id, MAX_NUM_FRIENDS).Iter()

	friend_list := make([]*api.FriendDTO, 0, 10)

	var friend_id int64
	var friend_name string
	var picture_digest []byte

	for iter.Scan(&friend_id, &friend_name, &picture_digest) {

		friendDTO := &api.FriendDTO{
			UserId:        friend_id,
			Name:          friend_name,
			PictureDigest: picture_digest,
		}

		friend_list = append(friend_list, friendDTO)
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadFriends:", err)
		return nil, convErr(err)
	}

	return friend_list, nil
}

func (d *FriendDAO) getFriends(user_id int64, friends_id ...int64) ([]*api.FriendDTO, error) {

	checkSession(d.session)

	stmt := `SELECT friend_id, friend_name, picture_digest FROM friends_by_user
						WHERE user_id = ? AND friend_id IN (` + utils.GenParams(len(friends_id)) + `) LIMIT ?`

	values := &utils.QueryValues{}
	values.AddValue(user_id)
	values.AddArrayInt64(friends_id)
	values.AddValue(MAX_NUM_FRIENDS)

	iter := d.session.Query(stmt, values.Params...).Iter()

	friend_list := make([]*api.FriendDTO, 0, 10)

	var friend_id int64
	var friend_name string
	var picture_digest []byte

	for iter.Scan(&friend_id, &friend_name, &picture_digest) {

		friendDTO := &api.FriendDTO{
			UserId:        friend_id,
			Name:          friend_name,
			PictureDigest: picture_digest,
		}

		friend_list = append(friend_list, friendDTO)
	}

	if err := iter.Close(); err != nil {
		log.Println("getFriends:", err)
		return nil, convErr(err)
	}

	return friend_list, nil
}

// Load members from database and store them into groups. If group owner mismatches
// user_id, it's ignored.
func (dao *FriendDAO) loadMembersIntoGroups(user_id int64, groups []*api.GroupDTO) error {

	checkSession(dao.session)

	// Build index (group_id => slice position) and groups ids
	index := make(map[int32]int)
	groups_ids := make([]int32, 0, len(groups))

	for i, group := range groups {
		index[group.Id] = i
		groups_ids = append(groups_ids, group.Id)
	}

	stmt := `SELECT group_id, friend_id FROM friends_by_group
		WHERE user_id = ? AND group_id IN (` + utils.GenParams(len(groups_ids)) + `)`

	var group_id int32
	var friend_id int64

	values := &utils.QueryValues{}
	values.AddValue(user_id)
	values.AddArrayInt32(groups_ids)

	iter := dao.session.Query(stmt, values.Params...).Iter()

	for iter.Scan(&group_id, &friend_id) {
		pos := index[group_id]
		groups[pos].Members = append(groups[pos].Members, friend_id)
	}

	if err := iter.Close(); err != nil {
		log.Println("loadMembersIntoGroups:", err)
		return convErr(err)
	}

	return nil
}

func (dao *FriendDAO) computeGroupSize(user_id int64, group_id int32) (int32, error) {

	checkSession(dao.session)

	stmt := `SELECT COUNT(*) AS group_size FROM friends_by_group
		WHERE user_id = ? AND group_id = ?`

	var group_size int32

	err := dao.session.Query(stmt, user_id, group_id).Scan(&group_size)
	if err != nil {
		return -1, convErr(err)
	}

	return group_size, nil
}

func (dao *FriendDAO) setGroupSize(user_id int64, group_id int32, group_size int32) error {
	checkSession(dao.session)
	stmt := `UPDATE groups_by_user SET group_size = ? WHERE user_id = ? AND group_id = ?`
	err := dao.session.Query(stmt, group_size, user_id, group_id).Exec()
	return convErr(err)
}

func (dao *FriendDAO) updateGroupSize(user_id int64, group_id int32) error {
	size, err := dao.computeGroupSize(user_id, group_id)
	if err != nil {
		return err
	}
	return dao.setGroupSize(user_id, group_id, size)
}
