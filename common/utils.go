package common

import (
	proto "areyouin/protocol"
	"github.com/gocql/gocql"
	"time"
)

func GetCurrentTimeMillis() int64 {
	return time.Now().UTC().UnixNano() / int64(time.Millisecond)
}

func ClearUserAccounts(session *gocql.Session) {
	session.Query(`TRUNCATE user_facebook_credentials`).Exec()
	session.Query(`TRUNCATE user_email_credentials`).Exec()
	session.Query(`TRUNCATE user_account`).Exec()
}

func ClearEvents(session *gocql.Session) {
	session.Query(`TRUNCATE event`).Exec()
	session.Query(`TRUNCATE event_participants`).Exec()
	session.Query(`TRUNCATE user_events`).Exec()
}

func CreateParticipantsFromFriends(author_id uint64, friends []*proto.Friend) []*proto.EventParticipant {

	result := make([]*proto.EventParticipant, 0, len(friends))

	if len(friends) > 0 {

		for _, f := range friends {
			result = append(result, &proto.EventParticipant{
				UserId:    f.UserId,
				Name:      f.Name,
				Response:  proto.AttendanceResponse_NO_RESPONSE,
				Delivered: proto.MessageStatus_NO_DELIVERED,
			})
		}
	}

	return result
}
