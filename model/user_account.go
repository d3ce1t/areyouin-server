package model

import (
	"peeple/areyouin/api"
	"peeple/areyouin/idgen"
	"peeple/areyouin/utils"
	"strings"

	"github.com/twinj/uuid"
)

const (
	UserPasswordMinLength = 5
	UserPasswordMaxLength = 50
	UserNameMinLength     = 3
	UserNameMaxLength     = 50
	UserPictureMaxWidth   = 512
	UserPictureMaxHeight  = 512
)

type UserAccount struct {
	id             int64 // AreYouIN ID
	name           string
	email          string // Contact e-mail
	emailVerified  bool
	phone          string
	phoneVerified  bool
	pictureDigest  []byte
	iidToken       IIDToken // Instance ID token
	lastConnection int64
	authToken      string
	emailCred      *EmailCredential
	fbCred         *FBCredential
	createdDate    int64

	// Indicate if this object has a copy in database. For instance,
	// an user loaded from db will have isPersisted set.
	isPersisted bool
}

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
		salt := utils.NewRandomSalt32()
		hashedPassword := utils.HashPasswordWithSalt(password, salt)
		user.emailCred = &EmailCredential{
			Email:    email,
			Salt:     salt,
			Password: hashedPassword,
		}
	}

	return user, nil
}

func newUserFromDTO(dto *api.UserDTO) *UserAccount {
	return &UserAccount{
		id:            dto.Id,
		name:          dto.Name,
		email:         dto.Email,
		emailVerified: dto.EmailVerified,
		pictureDigest: dto.PictureDigest,
		iidToken: IIDToken{
			token:   dto.IidToken.Token,
			version: dto.IidToken.Version,
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

func newUserListFromDTO(dtos []*api.UserDTO) []*UserAccount {
	results := make([]*UserAccount, 0, len(dtos))
	for _, userDTO := range dtos {
		results = append(results, newUserFromDTO(userDTO))
	}
	return results
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
	return u.iidToken
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

func (u *UserAccount) IsZero() bool {
	return u.id == 0 && u.name == "" && u.email == "" && !u.emailVerified &&
		u.phone == "" && !u.phoneVerified && u.pictureDigest == nil &&
		u.iidToken == IIDToken{} && u.lastConnection == 0 && u.authToken == "" &&
		u.emailCred == nil && u.fbCred == nil && u.createdDate == 0
}

func (u *UserAccount) Clone() *UserAccount {
	copy := new(UserAccount)
	*copy = *u
	if u.fbCred != nil {
		copy.fbCred = new(FBCredential)
		*(copy.fbCred) = *(u.fbCred)
	}
	if u.emailCred != nil {
		copy.emailCred = new(EmailCredential)
		*(copy.emailCred) = (*u.emailCred)
	}
	return copy
}

func (u *UserAccount) AsParticipant() *Participant {
	return NewParticipant(u.id, u.name, api.AttendanceResponse_NO_RESPONSE,
		api.InvitationStatus_SERVER_DELIVERED)
}

func (u *UserAccount) AsFriend() *Friend {
	return NewFriend(u.id, u.name, u.pictureDigest)
}

func (u *UserAccount) AsDTO() *api.UserDTO {

	userDTO := &api.UserDTO{
		Id:            u.id,
		AuthToken:     u.authToken,
		Email:         u.email,
		EmailVerified: u.emailVerified,
		Name:          u.name,
		IidToken:      *u.iidToken.AsDTO(),
		LastConn:      u.lastConnection,
		CreatedDate:   u.createdDate,
		PictureDigest: u.pictureDigest,
	}

	if u.fbCred != nil {
		userDTO.FbId = u.fbCred.FbId
		userDTO.FbToken = u.fbCred.Token
	}

	if u.emailCred != nil {
		userDTO.Password = u.emailCred.Password
		userDTO.Salt = u.emailCred.Salt
	}

	return userDTO
}

func validateUserData(name string, email string, password string, fbId string,
	fbToken string) error {

	// Name length
	if len(name) < UserNameMinLength || len(name) > UserNameMaxLength {
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
	if password == "" || len(password) < UserPasswordMinLength || len(password) > UserPasswordMaxLength {
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
	token string
	// Protocol Version stored when IIDtoken was received
	version  int
	platform string
}

func NewIIDToken(token string, version int, platform string) *IIDToken {
	return &IIDToken{
		token:    token,
		version:  version,
		platform: platform}
}

func newIIDTokenFromDTO(dto *api.IIDTokenDTO) *IIDToken {
	return &IIDToken{
		token:    dto.Token,
		version:  dto.Version,
		platform: dto.Platform,
	}
}

func (t *IIDToken) Token() string {
	return t.token
}

func (t *IIDToken) Version() int {
	return t.version
}

func (t *IIDToken) Platform() string {
	return t.platform
}

func (t IIDToken) AsDTO() *api.IIDTokenDTO {
	return &api.IIDTokenDTO{
		Token:    t.token,
		Version:  t.version,
		Platform: t.platform,
	}
}
