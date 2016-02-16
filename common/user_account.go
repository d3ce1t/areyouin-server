package common

import (
	"errors"
	"github.com/twinj/uuid"
	"regexp"
	"strings"
)

var validEmail = regexp.MustCompile(`\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3}`)

var (
	ErrInvalidEmail    = errors.New("invalid e-mail address")
	ErrInvalidName     = errors.New("invalid user name")
	ErrInvalidPassword = errors.New("password is too short")
	ErrNoCredentials   = errors.New("no credentials")
)

func NewEmptyUserAccount() *UserAccount {
	user := &UserAccount{}
	return user
}

// TODO: Password should always be hashed
func NewUserAccount(id uint64, name string, email string, password string, phone string, fbid string, fbtoken string) *UserAccount {

	user := &UserAccount{
		Id:          id,
		Name:        name,
		Email:       strings.ToLower(email),
		Password:    password,
		phone:       phone,
		Fbid:        fbid,
		Fbtoken:     fbtoken,
		AuthToken:   uuid.NewV4(),
		CreatedDate: GetCurrentTimeMillis()}

	return user
}

func (user *UserAccount) GetName() string {
	return user.Name
}

func (user *UserAccount) GetUserId() uint64 {
	return user.Id
}

// A valid user account always has an id, name and email, and at least one
// credential method, namely e-mail and password. or facebook
func (user *UserAccount) IsValid() (bool, error) {

	if user.Id == 0 || len(user.Name) < 3 {
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

func IsValidEmail(email string) bool {
	return validEmail.MatchString(email)
}

type UserAccount struct {
	Id             uint64 // AreYouIN ID
	AuthToken      uuid.UUID
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
	LastConnection int64
	CreatedDate    int64
	Picture        []byte
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
