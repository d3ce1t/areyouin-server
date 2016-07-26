package model

import (
	"peeple/areyouin/api"
)

type Friend struct {
	id            int64
	name          string
	pictureDigest []byte
}

func NewFriend(id int64, name string, pictureDigest []byte) *Friend {
	return &Friend{
		id:            id,
		name:          name,
		pictureDigest: pictureDigest,
	}
}

func NewFriendFromDTO(dto *api.FriendDTO) *Friend {
	return &Friend{
		id:            dto.UserId,
		name:          dto.Name,
		pictureDigest: dto.PictureDigest,
	}
}

func (f *Friend) Id() int64 {
	return f.id
}

func (f *Friend) Name() string {
	return f.name
}

func (f *Friend) PictureDigest() []byte {
	return f.pictureDigest
}

func (f *Friend) AsDTO() *api.FriendDTO {
	return &api.FriendDTO{
		UserId:        f.Id(),
		Name:          f.Name(),
		PictureDigest: f.PictureDigest(),
	}
}

type Group struct {
	id      int32
	name    string
	size    int
	members []int64
}

func NewGroup(id int32, name string, size int32) *Group {
	return &Group{
		id:   id,
		name: name,
		size: int(size),
	}
}

func (g *Group) Id() int32 {
	return g.id
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Size() int {
	return len(g.members)
}

/*func (g *Group) AddMember(userId int64) {
	g.pbg.Members = append(g.pbg.Members, userId)
}*/

/*func (g *Group) Members() []int64 {
	return g.pbg.Members
}*/

type FriendRequest struct {
	friendId    int64
	fromUser    int64
	name        string
	email       string
	createdDate int64
}

func NewFriendRequest(toUser int64, fromUser int64, name string, email string, createdDate int64) *FriendRequest {
	return &FriendRequest{
		friendId:    toUser,
		fromUser:    fromUser,
		name:        name,
		email:       email,
		createdDate: createdDate,
	}
}

/*type UserFriend interface {
	GetName() string
	GetUserId() int64
	GetPictureDigest() []byte
}*/

/*func (f *Friend) GetName() string {
	return f.Name
}

func (f *Friend) GetUserId() int64 {
	return f.UserId
}

func (f *Friend) GetPictureDigest() []byte {
	return f.PictureDigest
}*/
