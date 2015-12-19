package common

import (
	"github.com/twinj/uuid"
	proto "peeple/areyouin/protocol"
	"regexp"
)

var validEmail = regexp.MustCompile(`\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3}`)

func NewEmptyUserAccount() *UserAccount {
	user := &UserAccount{}
	return user
}

// TODO: Password should always be hashed
func NewUserAccount(id uint64, name string, email string, password string, phone string, fbid string, fbtoken string) *UserAccount {

	user := &UserAccount{
		Id:          id,
		Name:        name,
		Email:       email,
		Password:    password,
		phone:       phone,
		Fbid:        fbid,
		Fbtoken:     fbtoken,
		AuthToken:   uuid.NewV4(),
		CreatedDate: GetCurrentTimeMillis()}

	return user
}

func CheckUserAccount(user *UserAccount) bool {

	// A valid user account always has an id, name and email
	if user.Id == 0 || len(user.Name) < 3 || user.Email == "" || !IsValidEmail(user.Email) {
		return false
	}

	// Check if there is at least one credential
	valid := false

	if user.Password == "" {
		valid = user.HasFacebookCredentials()
	} else {
		valid = len(user.Password) >= 5
	}

	return valid
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
	LastConnection int64
	CreatedDate    int64
	//friends         map[uint64]*Friend
	//udb   *UsersDatabase // Database the user belongs to
	//inbox *Inbox
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

func (ua *UserAccount) AsFriend() *proto.Friend {
	return &proto.Friend{UserId: ua.Id, Name: ua.Name}
}

func (ua *UserAccount) AsParticipant() *proto.EventParticipant {
	participant := &proto.EventParticipant{
		UserId:    ua.Id,
		Name:      ua.Name,
		Response:  proto.AttendanceResponse_NO_RESPONSE,
		Delivered: proto.MessageStatus_NO_DELIVERED,
	}
	return participant
}
