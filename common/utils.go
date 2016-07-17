package common

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"image"
	"image/jpeg"
	"github.com/disintegration/imaging"
	"log"
	"math/big"
	"time"
	"github.com/gocql/gocql"
)

const (
	IMAGE_MDPI                 = 160              // 160dpi
	IMAGE_HDPI                 = 1.5 * IMAGE_MDPI // 240dpi
	IMAGE_XHDPI                = 2 * IMAGE_MDPI   // 320dpi
	IMAGE_XXHDPI               = 3 * IMAGE_MDPI   // 480dpi
	IMAGE_XXXHDPI              = 4 * IMAGE_MDPI   // 640dpi
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

func (self *QueryValues) AddArrayInt64(array []int64) {
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

func GenValues(values []int64) []interface{} {

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

func CreateParticipantsFromFriends(author_id int64, friends []*Friend) map[int64]*EventParticipant {

	result := make(map[int64]*EventParticipant)

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

func GetParticipantsIdSlice(participants map[int64]*EventParticipant) []int64 {
	result := make([]int64, 0, len(participants))
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

// Creates picture thumbnails for every supported dpi. Thumbnails size are
// (size_xy, size_xy)*scale_factor. Thumbnails are returned as byte slide
// JPEG encoded.
func CreateThumbnails(srcImage image.Image, size_xy int, forDpi []int32) (map[int32][]byte, error) {

	// Create thumbnails for distinct sizes
	thumbnails := make(map[int32][]byte)

	for _, dpi := range forDpi {
		// Compute size
		size := float32(size_xy) * (float32(dpi) / float32(IMAGE_MDPI))
		// Resize and crop image to fill size x size area
		dstImage := imaging.Thumbnail(srcImage, int(size), int(size), imaging.Lanczos)
		// Encode
		bytes := &bytes.Buffer{}
		err := jpeg.Encode(bytes, dstImage, nil)
		if err != nil {
			return nil, err
		}
		thumbnails[dpi] = bytes.Bytes()
	}

	return thumbnails, nil
}

func ResizeImage(picture image.Image, width int) ([]byte, error) {
	resize_image := imaging.Resize(picture, width, 0, imaging.Lanczos)
	bytes := &bytes.Buffer{}
	err := jpeg.Encode(bytes, resize_image, nil)
	if err != nil {
		return nil, err
	}
	return bytes.Bytes(), nil
}
