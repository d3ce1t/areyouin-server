package model

import (
	"github.com/twinj/uuid"
	"peeple/areyouin/api"
	"peeple/areyouin/idgen"
	"peeple/areyouin/utils"
	"strings"
)

const (
	USER_PASSWORD_MIN_LENGTH   = 5
	USER_NAME_MIN_LENGTH       = 3
	USER_NAME_MAX_LENGTH       = 50
	PROFILE_PICTURE_MAX_WIDTH  = 512
	PROFILE_PICTURE_MAX_HEIGHT = 512
)

type UserAccount struct {
	id             int64 // AreYouIN ID
	name           string
	email          string // Contact e-mail
	emailVerified  bool
	phone          string
	phoneVerified  bool
	pictureDigest  []byte
	iidToken       *IIDToken // Instance ID token
	lastConnection int64
	authToken      string
	emailCred      *EmailCredential
	fbCred         *FBCredential
	createdDate    int64
}

/*func NewEmptyUserAccount() *UserAccount {
	return &UserAccount{}
}*/

func NewUserAccount(name string, email string, password string, phone string,
	fbId string, fbToken string) (*UserAccount, error) {

	err := validateUserData(name, email, password, fbId, fbToken)
	if err != nil {
		return nil, err
	}

	user := &UserAccount{
		id:          idgen.NewID(),
		name:        name,
		email:       strings.ToLower(email),
		phone:       phone,
		authToken:   uuid.NewV4().String(),
		createdDate: utils.GetCurrentTimeMillis(),
	}

	if fbId != "" && fbToken != "" {
		user.fbCred = &FBCredential{
			FbId:  fbId,
			Token: fbToken,
		}
	}

	if email != "" && password != "" {
		user.emailCred = &EmailCredential{
			Email:    email,
			Salt:     utils.NewRandomSalt32(),
			Password: utils.HashPasswordWithSalt(password, user.emailCred.Salt),
		}
	}

	return user, nil
}

func NewUserFromDTO(dto *api.UserDTO) *UserAccount {
	return &UserAccount{
		id:            dto.Id,
		name:          dto.Name,
		email:         dto.Email,
		emailVerified: dto.EmailVerified,
		pictureDigest: dto.PictureDigest,
		iidToken: &IIDToken{
			Token:   dto.IidToken.Token,
			Version: dto.IidToken.Version,
		},
		authToken: dto.AuthToken,
		emailCred: &EmailCredential{
			Email:    dto.Email,
			Password: dto.Password,
			Salt:     dto.Salt,
		},
		fbCred: &FBCredential{
			FbId:  dto.FbId,
			Token: dto.FbToken,
		},
		lastConnection: dto.LastConn,
		createdDate:    dto.CreatedDate,
	}
}

func (u *UserAccount) Id() int64 {
	return u.id
}

func (u *UserAccount) Name() string {
	return u.name
}

func (u *UserAccount) Email() string {
	return u.email
}

func (u *UserAccount) AuthToken() string {
	return u.authToken
}

func (u *UserAccount) FbId() string {
	if u.fbCred != nil {
		return u.fbCred.FbId
	} else {
		return ""
	}
}

func (u *UserAccount) FbToken() string {
	if u.fbCred != nil {
		return u.fbCred.Token
	} else {
		return ""
	}
}

func (u *UserAccount) PushToken() IIDToken {
	if u.iidToken != nil {
		return *u.iidToken
	} else {
		return IIDToken{}
	}
}

func (u *UserAccount) PictureDigest() []byte {
	return u.pictureDigest
}

func (u *UserAccount) CreatedDate() int64 {
	return u.createdDate
}

func (u *UserAccount) HasFacebook() bool {
	return u.fbCred != nil && u.fbCred.FbId != "" && u.fbCred.Token != ""
}

/*func (u *UserAccount) HasEmailCredentials() bool {
	result := false
	if u.emailCred.Email != "" && u.emailCred.Password  password != "" {
		result = true
	}
	return result
}*/

func (u *UserAccount) AsParticipant() *Participant {
	return NewParticipant(u.id, u.name, api.AttendanceResponse_NO_RESPONSE,
		api.InvitationStatus_NO_DELIVERED)
}

func (u *UserAccount) AsFriend() *Friend {
	return NewFriend(u.id, u.name, u.pictureDigest)
}

func (u *UserAccount) AsDTO() *api.UserDTO {
	return &api.UserDTO{
		Id:            u.id,
		AuthToken:     u.authToken,
		Email:         u.email,
		EmailVerified: u.emailVerified,
		Password:      u.emailCred.Password,
		Salt:          u.emailCred.Salt,
		Name:          u.name,
		FbId:          u.fbCred.FbId,
		FbToken:       u.fbCred.Token,
		IidToken:      u.iidToken.AsDTO(),
		LastConn:      u.lastConnection,
		CreatedDate:   u.createdDate,
		PictureDigest: u.pictureDigest,
	}
}

func validateUserData(name string, email string, password string, fbId string,
	fbToken string) error {

	// Name length
	if len(name) < USER_NAME_MIN_LENGTH || len(name) > USER_NAME_MAX_LENGTH {
		return ErrInvalidName
	}

	// Mandatory e-mail
	if !utils.IsValidEmail(email) {
		return ErrInvalidEmail
	}

	// Check credentials

	var hasFbCred bool
	var hasEmailCred bool

	if password != "" {
		if !isValidPassword(password) {
			return ErrInvalidPassword
		}
		hasEmailCred = true
	}

	if fbId != "" && fbToken != "" {
		hasFbCred = true
	}

	if !hasEmailCred && !hasFbCred {
		return ErrNoCredentials
	}

	return nil
}

func isValidPassword(password string) bool {
	if password == "" || len(password) < USER_PASSWORD_MIN_LENGTH {
		return false
	}
	return true
}

type EmailCredential struct {
	Email    string
	Password [32]byte
	Salt     [32]byte
}

type FBCredential struct {
	FbId  string
	Token string
}

type IIDToken struct {
	Token string
	// Protocol Version stored when IIDtoken was received
	Version int
}

func (t *IIDToken) AsDTO() api.IIDTokenDTO {
	return api.IIDTokenDTO{
		Token:   t.Token,
		Version: t.Version,
	}
}
