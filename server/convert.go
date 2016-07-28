package main

import (
	"peeple/areyouin/model"
	"peeple/areyouin/protocol/core"
)

func convUser2Net(user *model.UserAccount) *core.UserAccount {
	return &core.UserAccount{
		Name:          user.Name(),
		Email:         user.Email(),
		PictureDigest: user.PictureDigest(),
		FbId:          user.FbId(),
	}
}

func convEvent2Net(event *model.Event) *core.Event {

	netEvent := &core.Event{
		EventId:       event.Id(),
		AuthorId:      event.AuthorId(),
		AuthorName:    event.AuthorName(),
		StartDate:     event.StartDate(),
		EndDate:       event.EndDate(),
		Message:       event.Description(),
		NumAttendees:  int32(event.NumAttendees()),
		NumGuests:     int32(event.NumGuests()),
		CreatedDate:   event.CreatedDate(),
		InboxPosition: event.InboxPosition(),
		PictureDigest: event.PictureDigest(),
		State:         core.EventState(event.Status()),
		Participants:  make(map[int64]*core.EventParticipant),
	}

	for _, p := range event.Participants() {

		netEvent.Participants[p.Id()] = &core.EventParticipant{
			UserId:    p.Id(),
			Name:      p.Name(),
			Response:  core.AttendanceResponse(p.Response()),
			Delivered: core.InvitationStatus(p.InvitationStatus()),
		}
	}

	return netEvent
}

func convEventList2Net(eventList []*model.Event) []*core.Event {
	netEvents := make([]*core.Event, 0, len(eventList))
	for _, event := range eventList {
		netEvents = append(netEvents, convEvent2Net(event))
	}
	return netEvents
}

func convParticipant2Net(participant *model.Participant) *core.EventParticipant {
	return &core.EventParticipant{
		UserId:    participant.Id(),
		Name:      participant.Name(),
		Response:  core.AttendanceResponse(participant.Response()),
		Delivered: core.InvitationStatus(participant.InvitationStatus()),
	}
}

func convParticipantList2Net(pl map[int64]*model.Participant) map[int64]*core.EventParticipant {
	result := make(map[int64]*core.EventParticipant)
	for _, p := range pl {
		result[p.Id()] = convParticipant2Net(p)
	}
	return result
}

func convFriend2Net(friend *model.Friend) *core.Friend {
	return &core.Friend{
		UserId:        friend.Id(),
		Name:          friend.Name(),
		PictureDigest: friend.PictureDigest(),
	}
}

func convFriendList2Net(friendList []*model.Friend) []*core.Friend {
	result := make([]*core.Friend, 0, len(friendList))
	for _, f := range friendList {
		result = append(result, convFriend2Net(f))
	}
	return result
}

func convFriendRequest2Net(friendRequest *model.FriendRequest) *core.FriendRequest {
	return &core.FriendRequest{
		FriendId:    friendRequest.FromUser(),
		Name:        friendRequest.FromUserName(),
		Email:       friendRequest.FromUserEmail(),
		CreatedDate: friendRequest.CreatedDate(),
	}
}

func convFriendRequestList2Net(friendRequestList []*model.FriendRequest) []*core.FriendRequest {
	result := make([]*core.FriendRequest, 0, len(friendRequestList))
	for _, r := range friendRequestList {
		result = append(result, convFriendRequest2Net(r))
	}
	return result
}

func convGroup2Net(group *model.Group) *core.Group {
	return &core.Group{
		Id:      group.Id(),
		Name:    group.Name(),
		Size:    int32(group.Size()),
		Members: group.Members(),
	}
}

func convGroupList2Net(groups []*model.Group) []*core.Group {
	result := make([]*core.Group, 0, len(groups))
	for _, g := range groups {
		result = append(result, convGroup2Net(g))
	}
	return result
}
