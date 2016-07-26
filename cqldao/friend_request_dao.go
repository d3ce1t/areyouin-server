package cqldao

import (
	"github.com/gocql/gocql"
	"peeple/areyouin/api"
)

type FriendRequestDAO struct {
	session *GocqlSession
}

/*func (dao *FriendDAO) LoadFriendRequest(user_id int64, friend_id int64) (*FriendRequestDTO, error) {

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
	friendRequest := &FriendRequestDTO{
		FromUser:    user_id,
		ToUser:      friend_id,
		Name:        name,
		Email:       email,
		CreatedDate: created_date,
	}

	return friendRequest, nil
}*/

func (d *FriendRequestDAO) LoadAll(user_id int64) ([]*api.FriendRequestDTO, error) {

	if user_id == 0 {
		return nil, api.ErrNotFound
	}

	checkSession(d.session)

	stmt := `SELECT created_date, friend_id, name, email FROM friend_requests_received
		WHERE user_id = ?`

	iter := d.session.Query(stmt, user_id).Iter()

	requests := make([]*api.FriendRequestDTO, 0, 10)

	var created_date int64
	var friend_id int64
	var name string
	var email string

	for iter.Scan(&created_date, &friend_id, &name, &email) {
		requests = append(requests, &api.FriendRequestDTO{
			FromUser:    friend_id,
			ToUser:      user_id,
			Name:        name,
			Email:       email,
			CreatedDate: created_date,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, convErr(err)
	}

	return requests, nil
}

// Check if exists a friend request from user_id to friend_id. In other words, if user_id
// has already sent (or not) a friend request to friend_id.
func (d *FriendRequestDAO) Exist(toUser int64, fromUser int64) (bool, error) {

	checkSession(d.session)

	if toUser == 0 || fromUser == 0 {
		return false, api.ErrInvalidArg
	}

	stmt := `SELECT created_date FROM friend_requests_sent
	 	WHERE user_id = ? AND friend_id = ? LIMIT 1`

	var created_date int64

	if err := d.session.Query(stmt, toUser, fromUser).Scan(&created_date); err != nil {
		if err == gocql.ErrNotFound {
			return false, nil
		} else {
			return false, convErr(err)
		}
	}

	return true, nil
}

// Insert a friend request into database. This request means that some user (friend_id)
// wants to be friend of user_id
func (d *FriendRequestDAO) Insert(friendRequest *api.FriendRequestDTO) error {

	checkSession(d.session)

	if friendRequest.FromUser == 0 || friendRequest.ToUser == 0 {
		return api.ErrInvalidArg
	}

	stmt_insert_sent := `INSERT INTO friend_requests_sent (user_id, friend_id, created_date)
		VALUES (?, ?, ?)`

	stmt_insert_received := `INSERT INTO friend_requests_received (user_id, created_date, friend_id, name, email)
		VALUES (?, ?, ?, ?, ?)`

	batch := d.session.NewBatch(gocql.LoggedBatch)
	batch.Query(stmt_insert_sent, friendRequest.FromUser, friendRequest.ToUser,
		friendRequest.CreatedDate)
	batch.Query(stmt_insert_received, friendRequest.ToUser, friendRequest.CreatedDate,
		friendRequest.FromUser, friendRequest.Name, friendRequest.Email)

	return convErr(d.session.ExecuteBatch(batch))
}

func (d *FriendRequestDAO) Delete(friendRequest *api.FriendRequestDTO) error {

	checkSession(d.session)

	if friendRequest.FromUser == 0 || friendRequest.ToUser == 0 {
		return api.ErrInvalidArg
	}

	stmt_delete_sent := `DELETE FROM friend_requests_sent WHERE user_id = ? AND friend_id = ?`
	stmt_delete_received := `DELETE FROM friend_requests_received
		WHERE user_id = ? AND created_date = ? AND friend_id = ?`

	batch := d.session.NewBatch(gocql.LoggedBatch)
	batch.Query(stmt_delete_sent, friendRequest.FromUser, friendRequest.ToUser)
	batch.Query(stmt_delete_received, friendRequest.ToUser, friendRequest.CreatedDate,
		friendRequest.FromUser)

	return convErr(d.session.ExecuteBatch(batch))
}
