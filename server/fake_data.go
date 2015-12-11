package main

import (
	proto "areyouin/protocol"
	"time"
)

func initFakeUsers(server *Server) {

	user1 := NewUserAccount(server.GetNewUserID(), "User 1", "user1@foo.com", "12345", "", "", "")
	user2 := NewUserAccount(server.GetNewUserID(), "User 2", "user2@foo.com", "12345", "", "", "")
	user3 := NewUserAccount(server.GetNewUserID(), "User 3", "user3@foo.com", "12345", "", "", "")
	user4 := NewUserAccount(server.GetNewUserID(), "User 4", "user4@foo.com", "12345", "", "", "")
	user5 := NewUserAccount(server.GetNewUserID(), "User 5", "user5@foo.com", "12345", "", "", "")
	user6 := NewUserAccount(server.GetNewUserID(), "User 6", "user6@foo.com", "12345", "", "", "")
	user7 := NewUserAccount(server.GetNewUserID(), "User 7", "user7@foo.com", "12345", "", "", "")
	user8 := NewUserAccount(server.GetNewUserID(), "User 8", "user8@foo.com", "12345", "", "", "")

	server.udb.Insert(user1)
	server.udb.Insert(user2)
	server.udb.Insert(user3)
	server.udb.Insert(user4)
	server.udb.Insert(user5)
	server.udb.Insert(user6)
	server.udb.Insert(user7)
	server.udb.Insert(user8)

	user1.AddFriend(user2.id)
	user1.AddFriend(user3.id)
	user1.AddFriend(user4.id)
	user1.AddFriend(user5.id)
	user1.AddFriend(user6.id)
	user1.AddFriend(user7.id)
	user1.AddFriend(user8.id)

	user2.AddFriend(user1.id)
	user2.AddFriend(user3.id)
	user2.AddFriend(user4.id)

	user3.AddFriend(user1.id)
	user3.AddFriend(user2.id)
	user3.AddFriend(user4.id)

	user4.AddFriend(user1.id)
	user4.AddFriend(user2.id)
	user4.AddFriend(user3.id)
}

func initFakeEvents(server *Server) {

	author, _ := server.udb.GetByEmail("user1@foo.com")

	// Create event with user1 as author
	event1 := &proto.Event{
		EventId:            server.GetNewUserID(), // Maybe a bottleneck here
		AuthorId:           author.id,
		AuthorName:         author.name,
		CreationDate:       time.Now().UTC().Unix(), // Seconds
		StartDate:          time.Now().UTC().Unix(),
		EndDate:            time.Now().UTC().Unix(),
		Message:            "test",
		IsPublic:           false,
		NumberParticipants: 1, // The own author
	}

	// Add all author's friends as participants
	for _, friend := range author.GetAllFriends() {
		participant := &proto.EventParticipant{
			UserId:    friend.id,
			Name:      friend.name,
			Response:  proto.AttendanceResponse_NO_RESPONSE,
			Delivered: proto.MessageStatus_NO_DELIVERED,
		}
		event1.Participants = append(event1.Participants, participant)
	}

	// Add author as another participant
	event1.Participants = append(event1.Participants, &proto.EventParticipant{
		UserId:    author.id,
		Name:      author.name,
		Response:  proto.AttendanceResponse_ASSIST,
		Delivered: proto.MessageStatus_NO_DELIVERED,
	})

	event1.NumberParticipants = uint32(len(event1.Participants))

	if ok := server.edb.Insert(event1); ok { // Insert is not thread-safe
		server.ds.Submit(event1)
	}

	time.Sleep(time.Second)
}
