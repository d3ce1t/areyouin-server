package common

import (
	"crypto/rand"
	"crypto/sha256"
	"github.com/gocql/gocql"
	"log"
	"time"
)

var (
	EMPTY_ARRAY_32B = [32]byte{}
)

// Get current time in millis
func GetCurrentTimeMillis() int64 {
	return TimeToMillis(time.Now())
}

// Return current time in millis with seconds precision
func GetCurrentTimeSeconds() int64 {
	return TimeToSeconds(time.Now()) * 1000
}

// Return time as millis
func TimeToMillis(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// Return time as seconds
func TimeToSeconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Second)
}

func UnixMillisToTime(timestamp int64) time.Time {
	seconds := timestamp / 1000
	millis := timestamp % 1000
	return time.Unix(seconds, millis*int64(time.Millisecond))
}

func ClearUserAccounts(session *gocql.Session) {
	session.Query(`TRUNCATE user_facebook_credentials`).Exec()
	session.Query(`TRUNCATE user_email_credentials`).Exec()
	session.Query(`TRUNCATE user_account`).Exec()
	session.Query(`TRUNCATE user_friends`).Exec()
}

func ClearEvents(session *gocql.Session) {
	session.Query(`TRUNCATE event`).Exec()
	//session.Query(`TRUNCATE event_participants`).Exec()
	session.Query(`TRUNCATE user_events`).Exec()
}

func CreateParticipantsFromFriends(author_id uint64, friends []*Friend) []*EventParticipant {

	result := make([]*EventParticipant, 0, len(friends))

	if len(friends) > 0 {

		for _, f := range friends {
			result = append(result, &EventParticipant{
				UserId:    f.UserId,
				Name:      f.Name,
				Response:  AttendanceResponse_NO_RESPONSE,
				Delivered: MessageStatus_NO_DELIVERED,
			})
		}
	}

	return result
}

/*func Log(message string) {
	fmt.Println(message)
}*/

func NewRandomSalt32() (salt [32]byte, err error) {
	_, err = rand.Read(salt[:])
	if err != nil {
		log.Println("NewRandomSalt32() error:", err)
	}
	return
}

func HashPasswordWithSalt(password string, salt [32]byte) [32]byte {
	data := []byte(password)
	data = append(data, salt[:]...)
	return sha256.Sum256(data)
}
