package model

import (
	"peeple/areyouin/api"
)

type AuthCredential struct {
	UserId int64
	Token  string
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
