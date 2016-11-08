package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"image"
	"image/jpeg"
	"math/big"
	"regexp"
	"time"

	"github.com/disintegration/imaging"
)

const (
	IMAGE_MDPI    = 160              // 160dpi
	IMAGE_HDPI    = 1.5 * IMAGE_MDPI // 240dpi
	IMAGE_XHDPI   = 2 * IMAGE_MDPI   // 320dpi
	IMAGE_XXHDPI  = 3 * IMAGE_MDPI   // 480dpi
	IMAGE_XXXHDPI = 4 * IMAGE_MDPI   // 640dpi
)

var (
	validEmail = regexp.MustCompile(`\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3}`)
)

func IsValidEmail(email string) bool {

	if email == "" || len(email) > 254 {
		return false
	}

	// Golang regex MatchString tries to match the left-most substring, not the whole
	// string. So this is a workaround to check string matching
	// --- start work around ---
	match := validEmail.FindString(email)
	result := false

	if match != "" {
		result = match == email
	}
	// --- end work around ---

	return result
}

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
func GetCurrentTimeUTC() time.Time {
	return time.Now().UTC()
}

func GetCurrentTimeMillis() int64 {
	return TimeToMillis(time.Now())
}

// Return time as millis
func TimeToMillis(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func MillisToTimeUTC(timestamp int64) time.Time {

	if timestamp > 0 {
		seconds := timestamp / 1000
		millis := timestamp % 1000
		return time.Unix(seconds, millis*int64(time.Millisecond)).UTC()
	}

	return time.Time{}
}

func CreateDate(year, month, day, hour, min int) time.Time {
	return time.Date(year, time.Month(month), day, hour, min, 0, 0, time.UTC)
}

func RandUint16() (uint16, error) {
	v, err := rand.Int(rand.Reader, big.NewInt(65536))
	return uint16(v.Int64()), err
}

func NewRandomSalt32() [32]byte {
	var salt [32]byte
	_, err := rand.Read(salt[:])
	if err != nil {
		panic(err)
	}
	return salt
}

func HashPasswordWithSalt(password string, salt [32]byte) [32]byte {
	data := []byte(password)
	data = append(data, salt[:]...)
	return sha256.Sum256(data)
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func MinUint(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}

func MaxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func MinDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func MaxDuration(a, b time.Duration) time.Duration {
	if a > b {
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
