package model

import (
	"log"
	"peeple/areyouin/api"
	"time"
)

type ParticipantAdder interface {
	AddUserAccount(u *UserAccount)
	AddFriend(f *Friend)
	AddParticipant(p *Participant)
	AddUserID(UID int64)
}

type participantListCreator struct {
	eventManager *EventManager
	participants map[int64]interface{}
	eventID      int64
	ownerID      int64
	timestamp    int64
}

func (m *EventManager) newParticipantListCreator() *participantListCreator {
	return &participantListCreator{
		eventManager: m,
		participants: make(map[int64]interface{}),
		timestamp:    time.Now().UnixNano() / 1000,
	}
}

func (b *participantListCreator) SetOwner(ownerID int64) {
	b.ownerID = ownerID
}

func (b *participantListCreator) SetEventID(eventID int64) {
	b.eventID = eventID
}

func (b *participantListCreator) SetTimestamp(t int64) {
	b.timestamp = t
}

func (b *participantListCreator) AddUserAccount(u *UserAccount) {
	b.participants[u.id] = u.AsParticipant()
}

func (b *participantListCreator) AddFriend(f *Friend) {
	b.participants[f.id] = NewParticipant(f.id, f.name, api.AttendanceResponse_NO_RESPONSE,
		api.InvitationStatus_SERVER_DELIVERED)
}

func (b *participantListCreator) AddParticipant(p *Participant) {
	b.participants[p.id] = p.Clone()
}

func (b *participantListCreator) AddUserID(UID int64) {
	b.participants[UID] = UID
}

func (b *participantListCreator) Len() int {
	return len(b.participants)
}

func (b *participantListCreator) Build() (map[int64]*Participant, error) {

	if b.ownerID == 0 {
		return nil, ErrMissingArgument
	}

	result := make(map[int64]*Participant)

	for _, v := range b.participants {

		var participant *Participant

		if pID, ok := v.(int64); ok {

			user, err := b.eventManager.parent.Accounts.GetUserAccount(pID)
			if err == api.ErrNotFound {
				log.Printf("* WARNING: CREATE PARTICIPANT LIST -> USER %v NOT FOUND\n", pID)
				continue
			} else if err != nil {
				return nil, err
			}

			participant = user.AsParticipant()

		} else {
			participant = v.(*Participant)
		}

		isFriend, err := b.eventManager.parent.Friends.IsFriend(participant.id, b.ownerID)

		if err != nil {
			return nil, err
		} else if !isFriend {
			log.Printf("* WARNING: CREATE PARTICIPANT LIST -> USER %v TRIED TO ADD USER %v BUT THEY ARE NOT FRIENDS\n",
				b.ownerID, participant.id)
			continue
		}

		participant.eventID = b.eventID
		participant.responseTS = b.timestamp
		participant.statusTS = b.timestamp
		result[participant.id] = participant
	}

	if len(result) == 0 {
		return nil, ErrParticipantsRequired
	}

	return result, nil
}
