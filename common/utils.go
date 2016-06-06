package common

import (
	"crypto/rand"
	"crypto/sha256"
	"log"
	"math/big"
	"time"

	"github.com/gocql/gocql"
)

var (
	EMPTY_ARRAY_32B = [32]byte{}
)

type QueryValues struct {
	Params []interface{}
}

func (self *QueryValues) AddValue(val interface{}) {
	self.Params = append(self.Params, val)
}

func (self *QueryValues) AddArrayInt32(array []int32) {
	for _, val := range array {
		self.Params = append(self.Params, val)
	}
}

func (self *QueryValues) AddArrayUint32(array []uint32) {
	for _, val := range array {
		self.Params = append(self.Params, val)
	}
}

func (self *QueryValues) AddArrayUint32AsInt32(array []uint32) {
	for _, val := range array {
		self.Params = append(self.Params, int32(val))
	}
}

func (self *QueryValues) AddArrayUint64AsInt64(array []uint64) {
	for _, val := range array {
		self.Params = append(self.Params, int64(val))
	}
}

func (self *QueryValues) AddArrayUint64(array []uint64) {
	for _, val := range array {
		self.Params = append(self.Params, val)
	}
}

func GenParams(size int) string {

	if size == 0 {
		return ""
	}

	result := "?"
	for i := 1; i < size; i++ {
		result += ", ?"
	}
	return result
}

func GenValues(values []uint64) []interface{} {

	result := make([]interface{}, 0, len(values))

	for _, val := range values {
		result = append(result, val)
	}

	return result
}

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

func CreateParticipantsFromFriends(author_id uint64, friends []*Friend) map[uint64]*EventParticipant {

	result := make(map[uint64]*EventParticipant)

	if len(friends) > 0 {

		for _, f := range friends {
			result[f.UserId] = &EventParticipant{
				UserId:    f.UserId,
				Name:      f.Name,
				Response:  AttendanceResponse_NO_RESPONSE,
				Delivered: MessageStatus_NO_DELIVERED,
			}
		}
	}

	return result
}

func GetParticipantsIdSlice(participants map[uint64]*EventParticipant) []uint64 {
	result := make([]uint64, 0, len(participants))
	for _, p := range participants {
		result = append(result, p.UserId)
	}
	return result
}

/*func Log(message string) {
	fmt.Println(message)
}*/

func RandUint16() (uint16, error) {
	v, err := rand.Int(rand.Reader, big.NewInt(65536))
	return uint16(v.Int64()), err
}

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

func MinUint32(a, b uint32) uint32 {
	if a <= b {
		return a
	}
	return b
}
