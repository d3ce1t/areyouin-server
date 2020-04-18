package model

import (
	"github.com/d3ce1t/areyouin-server/api"
)

func newAccesToken(userID int64, token string) *AccessToken {
	return &AccessToken{userID: userID, token: token}
}

func newAccessTokenFromDTO(dto *api.AccessTokenDTO) *AccessToken {
	return &AccessToken{userID: dto.UserId, token: dto.Token}
}

type AccessToken struct {
	userID int64
	token  string
}

func (t *AccessToken) UserID() int64 {
	return t.userID
}

func (t *AccessToken) Token() string {
	return t.token
}

func (t *AccessToken) AsDTO() *api.AccessTokenDTO {
	return &api.AccessTokenDTO{
		UserId: t.userID,
		Token:  t.token,
	}
}

type Picture struct {
	RawData []byte
	Digest  []byte
}

func (p *Picture) AsDTO() *api.PictureDTO {
	return &api.PictureDTO{
		RawData: p.RawData,
		Digest:  p.Digest,
	}
}
