package common

import (
	"errors"
	"regexp"
	"strings"
	"github.com/twinj/uuid"
	"peeple/areyouin/idgen"
)

var validEmail = regexp.MustCompile(`\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3}`)

var (
	ErrInvalidEmail    = errors.New("invalid e-mail address")
	ErrInvalidName     = errors.New("invalid user name")
	ErrInvalidPassword = errors.New("password is too short")
	ErrNoCredentials   = errors.New("no credentials")
)

const (
	USER_NAME_MIN_LENGTH         = 3
	USER_NAME_MAX_LENGTH         = 50
	PROFILE_PICTURE_MAX_WIDTH  = 512
	PROFILE_PICTURE_MAX_HEIGHT = 512
)

func NewEmptyUserAccount() *UserAccount {
	user := &UserAccount{}
	return user
}

// TODO: Password should always be hashed
func NewUserAccount(name string, email string, password string, phone string, fbid string, fbtoken string) *UserAccount {

	user := &UserAccount{
		Id:          idgen.NewID(),
		Name:        name,
		Email:       strings.ToLower(email),
		Password:    password,
		phone:       phone,
		Fbid:        fbid,
		Fbtoken:     fbtoken,
		AuthToken:   uuid.NewV4().String(),
		CreatedDate: GetCurrentTimeMillis()}

	return user
}

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

type UserAccount struct {
	Id             int64 // AreYouIN ID
	AuthToken      string
	Email          string
	EmailVerified  bool
	Password       string
	salt           [32]byte
	Name           string
	phone          string
	phone_verified bool
	Fbid           string // Facebook ID
	Fbtoken        string // Facebook User Access token
	IIDtoken       string // Instance ID token
	NetworkVersion int // Protocol Version stored when IIDtoken was received
	LastConnection int64
	CreatedDate    int64
	Picture        []byte
	PictureDigest  []byte
}

// A valid user account always has an id, name and email, and at least one
// credential method, namely e-mail and password. or facebook
func (user *UserAccount) IsValid() (bool, error) {

	if user.Id == 0 || len(user.Name) < USER_NAME_MIN_LENGTH || len(user.Name) > USER_NAME_MAX_LENGTH {
		return false, ErrInvalidName
	}

	if user.Email == "" || !IsValidEmail(user.Email) {
		return false, ErrInvalidEmail
	}

	// Check if there is at least one credential
	exist_credential := false

	if user.Password != "" {
		if len(user.Password) < 5 {
			return false, ErrInvalidPassword
		} else {
			exist_credential = true
		}
	} else {
		exist_credential = user.HasFacebookCredentials()
	}

	if !exist_credential {
		return false, ErrNoCredentials
	}

	return true, nil
}

func (user *UserAccount) GetName() string {
	return user.Name
}

func (user *UserAccount) GetUserId() int64 {
	return user.Id
}

func (user *UserAccount) GetPictureDigest() []byte {
	return user.PictureDigest
}

func (ua *UserAccount) HasFacebookCredentials() bool {
	result := false
	if ua.Fbid != "" && ua.Fbtoken != "" {
		result = true
	}
	return result
}

func (ua *UserAccount) HasEmailCredentials() bool {
	result := false
	if ua.Email != "" && ua.Password != "" {
		result = true
	}
	return result
}

func (ua *UserAccount) AsParticipant() *EventParticipant {
	participant := &EventParticipant{
		UserId:    ua.Id,
		Name:      ua.Name,
		Response:  AttendanceResponse_NO_RESPONSE,
		Delivered: MessageStatus_NO_DELIVERED,
	}
	return participant
}
