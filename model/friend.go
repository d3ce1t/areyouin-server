package model

import (
	"github.com/d3ce1t/areyouin-server/api"
	"github.com/d3ce1t/areyouin-server/utils"
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

func newFriendFromDTO(dto *api.FriendDTO) *Friend {
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

func newGroupFromDTO(dto *api.GroupDTO) *Group {
	group := NewGroup(dto.Id, dto.Name, dto.Size)
	for _, memberID := range dto.Members {
		group.members = append(group.members, memberID)
	}
	return group
}

func newGroupListFromDTO(dtos []*api.GroupDTO) []*Group {
	groupList := make([]*Group, 0, len(dtos))
	for _, groupDTO := range dtos {
		groupList = append(groupList, newGroupFromDTO(groupDTO))
	}
	return groupList
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

func (g *Group) Members() []int64 {
	copy := make([]int64, 0, len(g.members))
	for _, gid := range g.members {
		copy = append(copy, gid)
	}
	return copy
}

func (g *Group) Clone() *Group {
	copy := *g
	copy.members = make([]int64, 0, len(g.members))
	for _, gid := range g.members {
		copy.members = append(copy.members, gid)
	}
	return &copy
}

func (g *Group) AsDTO() *api.GroupDTO {

	dto := &api.GroupDTO{
		Id:   g.id,
		Name: g.name,
		Size: int32(g.size),
	}

	for _, friendID := range g.members {
		dto.Members = append(dto.Members, friendID)
	}

	return dto
}

type GroupBuilder struct {
	g Group
}

func NewGroupBuilder() *GroupBuilder {
	return &GroupBuilder{}
}

func (b *GroupBuilder) SetId(id int32) *GroupBuilder {
	b.g.id = id
	return b
}

func (b *GroupBuilder) SetName(name string) *GroupBuilder {
	b.g.name = name
	return b
}

func (b *GroupBuilder) AddMember(friendId int64) *GroupBuilder {
	b.g.members = append(b.g.members, friendId)
	b.g.size = len(b.g.members)
	return b
}

func (b *GroupBuilder) Build() *Group {
	return b.g.Clone()
}

type FriendRequest struct {
	toUser        int64
	fromUser      int64
	fromUserName  string
	fromUserEmail string
	createdDate   int64
}

func NewFriendRequest(toUser int64, fromUser int64, name string, email string) *FriendRequest {
	return &FriendRequest{
		toUser:        toUser,
		fromUser:      fromUser,
		fromUserName:  name,
		fromUserEmail: email,
		createdDate:   utils.GetCurrentTimeMillis(),
	}
}

func newFriendRequestFromDTO(dto *api.FriendRequestDTO) *FriendRequest {
	return &FriendRequest{
		toUser:        dto.ToUser,
		fromUser:      dto.FromUser,
		fromUserName:  dto.Name,
		fromUserEmail: dto.Email,
		createdDate:   dto.CreatedDate,
	}
}

func newFriendRequestListFromDTO(dtos []*api.FriendRequestDTO) []*FriendRequest {
	results := make([]*FriendRequest, 0, len(dtos))
	for _, friendRequestDTO := range dtos {
		results = append(results, newFriendRequestFromDTO(friendRequestDTO))
	}
	return results
}

func (r *FriendRequest) FromUser() int64 {
	return r.fromUser
}

func (r *FriendRequest) ToUser() int64 {
	return r.toUser
}

func (r *FriendRequest) FromUserName() string {
	return r.fromUserName
}

func (r *FriendRequest) FromUserEmail() string {
	return r.fromUserEmail
}

func (r *FriendRequest) CreatedDate() int64 {
	return r.createdDate
}

func (r *FriendRequest) AsDTO() *api.FriendRequestDTO {
	return &api.FriendRequestDTO{
		ToUser:      r.toUser,
		FromUser:    r.fromUser,
		Name:        r.fromUserName,
		Email:       r.fromUserEmail,
		CreatedDate: r.createdDate,
	}
}
