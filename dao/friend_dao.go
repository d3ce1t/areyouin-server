package dao

import (
	"github.com/gocql/gocql"
	"log"
	core "peeple/areyouin/common"
)

type FriendDAO struct {
	session *gocql.Session
}

func NewFriendDAO(session *gocql.Session) core.FriendDAO {
	return &FriendDAO{session: session}
}

func (dao *FriendDAO) LoadFriendsIndex(user_id uint64, group_id int32) (map[uint64]*core.Friend, error) {

	dao.checkSession()

	stmt := `SELECT friend_id, name, picture_digest FROM user_friends
						WHERE user_id = ? AND group_id = ? LIMIT ?`

	iter := dao.session.Query(stmt, user_id, group_id, MAX_NUM_FRIENDS).Iter()

	friend_map := make(map[uint64]*core.Friend)

	var friend_id uint64
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
		log.Println("LoadFriendsIndex:", err)
		return nil, err
	}

	return friend_map, nil
}

func (dao *FriendDAO) LoadFriends(user_id uint64, group_id int32) ([]*core.Friend, error) {

	dao.checkSession()

	stmt := `SELECT friend_id, name, picture_digest FROM user_friends
						WHERE user_id = ? AND group_id = ? LIMIT ?`

	iter := dao.session.Query(stmt, user_id, group_id, MAX_NUM_FRIENDS).Iter()

	friend_list := make([]*core.Friend, 0, 10)

	var friend_id uint64
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

// Since makeFriends() is bidirectional (adds the friend to user1 and user2). It can
// be assumed that if first user is friend of the second one, then second user must also
// have the first user in his/her friend list.
func (dao *FriendDAO) IsFriend(user_id uint64, other_user_id uint64) (bool, error) {

	dao.checkSession()

	stmt := `SELECT user_id FROM user_friends
		WHERE user_id = ? AND group_id = ? AND friend_id = ?`

	err := dao.session.Query(stmt, user_id, 0, other_user_id).Scan(nil)

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

	stmt := `INSERT INTO user_friends (user_id, group_id, group_name, friend_id, name, picture_digest)
							VALUES (?, ?, ?, ?, ?, ?)`

	batch.Query(stmt, user1.GetUserId(), 0, "All Friends", user2.GetUserId(), user2.GetName(), user2.GetPictureDigest())

	batch.Query(stmt, user2.GetUserId(), 0, "All Friends", user1.GetUserId(), user1.GetName(), user1.GetPictureDigest())

	return dao.session.ExecuteBatch(batch)
}

func (dao *FriendDAO) SetPictureDigest(user_id uint64, friend_id uint64, digest []byte) error {
	dao.checkSession()
	stmt := `UPDATE user_friends SET picture_digest = ?
						WHERE user_id = ? AND group_id = ? AND friend_id = ?`
	return dao.session.Query(stmt, digest, user_id, 0, friend_id).Exec()
}

func (dao *FriendDAO) DeleteFriendsGroup(user_id uint64, group_id int32) error {

	dao.checkSession()

	if user_id == 0 {
		return ErrInvalidArg
	}

	return dao.session.Query(`DELETE FROM user_friends WHERE user_id = ? AND group_id = ?`,
		user_id, group_id).Exec()
}

func (dao *FriendDAO) checkSession() {
	if dao.session == nil {
		panic(ErrNoSession)
	}
}
