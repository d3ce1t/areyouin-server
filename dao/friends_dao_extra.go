package dao

import (
	"log"
	core "peeple/areyouin/common"
)

func (dao *FriendDAO) getFriendsIdInGroup(user_id int64, group_id int32) ([]int64, error) {

	dao.checkSession()

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
		return nil, err
	}

	return ids_slice, nil
}

func (dao *FriendDAO) getAllFriends(user_id int64) ([]*core.Friend, error) {

	dao.checkSession()

	stmt := `SELECT friend_id, friend_name, picture_digest FROM friends_by_user
						WHERE user_id = ? LIMIT ?`

	iter := dao.session.Query(stmt, user_id, MAX_NUM_FRIENDS).Iter()

	friend_list := make([]*core.Friend, 0, 10)

	var friend_id int64
	var friend_name string
	var picture_digest []byte

	for iter.Scan(&friend_id, &friend_name, &picture_digest) {
		friend_list = append(friend_list, &core.Friend{
			UserId:        friend_id,
			Name:          friend_name,
			PictureDigest: picture_digest,
		})
	}

	if err := iter.Close(); err != nil {
		log.Println("LoadFriends:", err)
		return nil, err
	}

	return friend_list, nil
}

func (dao *FriendDAO) getFriends(user_id int64, friends_id ...int64) ([]*core.Friend, error) {

	dao.checkSession()

	stmt := `SELECT friend_id, friend_name, picture_digest FROM friends_by_user
						WHERE user_id = ? AND friend_id IN (` + core.GenParams(len(friends_id)) + `) LIMIT ?`

	values := &core.QueryValues{}
	values.AddValue(user_id)
	values.AddArrayInt64(friends_id)
	values.AddValue(MAX_NUM_FRIENDS)

	iter := dao.session.Query(stmt, values.Params...).Iter()

	friend_list := make([]*core.Friend, 0, 10)

	var friend_id int64
	var friend_name string
	var picture_digest []byte

	for iter.Scan(&friend_id, &friend_name, &picture_digest) {
		friend_list = append(friend_list, &core.Friend{
			UserId:        friend_id,
			Name:          friend_name,
			PictureDigest: picture_digest,
		})
	}

	if err := iter.Close(); err != nil {
		log.Println("getFriends:", err)
		return nil, err
	}

	return friend_list, nil
}

// Load members from database and store them into groups. If group owner mismatches
// user_id, it's ignored.
func (dao *FriendDAO) loadMembersIntoGroups(user_id int64, groups []*core.Group) error {

	dao.checkSession()

	// Build index (group_id => slice position) and groups ids
	index := make(map[int32]int)
	groups_ids := make([]int32, 0, len(groups))

	for i, group := range groups {
		index[group.Id] = i
		groups_ids = append(groups_ids, group.Id)
	}

	stmt := `SELECT group_id, friend_id FROM friends_by_group
		WHERE user_id = ? AND group_id IN (` + core.GenParams(len(groups_ids)) + `)`

	var group_id int32
	var friend_id int64

	values := &core.QueryValues{}
	values.AddValue(user_id)
	values.AddArrayInt32(groups_ids)

	iter := dao.session.Query(stmt, values.Params...).Iter()

	for iter.Scan(&group_id, &friend_id) {
		pos := index[group_id]
		groups[pos].Members = append(groups[pos].Members, friend_id)
	}

	if err := iter.Close(); err != nil {
		log.Println("loadMembersIntoGroups:", err)
		return err
	}

	return nil
}

func (dao *FriendDAO) computeGroupSize(user_id int64, group_id int32) (int32, error) {

	dao.checkSession()

	stmt := `SELECT COUNT(*) AS group_size FROM friends_by_group
		WHERE user_id = ? AND group_id = ?`

	var group_size int32

	err := dao.session.Query(stmt, user_id, group_id).Scan(&group_size)
	if err != nil {
		return -1, err
	}

	return group_size, nil
}

func (dao *FriendDAO) setGroupSize(user_id int64, group_id int32, group_size int32) error {
	dao.checkSession()
	stmt := `UPDATE groups_by_user SET group_size = ? WHERE user_id = ? AND group_id = ?`
	return dao.session.Query(stmt, group_size, user_id, group_id).Exec()
}

func (dao *FriendDAO) updateGroupSize(user_id int64, group_id int32) error {
	size, err := dao.computeGroupSize(user_id, group_id)
	if err != nil {
		return err
	}
	return dao.setGroupSize(user_id, group_id, size)
}
