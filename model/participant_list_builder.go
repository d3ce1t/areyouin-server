package model

import (
	"log"
	"peeple/areyouin/api"
)

type ParticipantAdder interface {
	AddUserAccount(u *UserAccount)
	AddParticipant(p *Participant)
	AddFriend(f *Friend)
	AddUserID(UID int64)
}

type participantListCreator struct {
	eventManager *EventManager
	participants map[int64]interface{}
	authorID     int64
}

func (m *EventManager) newParticipantListCreator() *participantListCreator {
	return &participantListCreator{
		eventManager: m,
		participants: make(map[int64]interface{}),
	}
}

func (b *participantListCreator) SetAuthor(authorID int64) {
	b.authorID = authorID
}

func (b *participantListCreator) AddUserAccount(u *UserAccount) {
	b.participants[u.id] = u.AsParticipant()
}

func (b *participantListCreator) AddParticipant(p *Participant) {
	b.participants[p.id] = p
}

func (b *participantListCreator) AddFriend(f *Friend) {
	b.participants[f.id] = NewParticipant(f.id, f.name, api.AttendanceResponse_NO_RESPONSE,
		api.InvitationStatus_SERVER_DELIVERED)
}

func (b *participantListCreator) AddUserID(UID int64) {
	b.participants[UID] = UID
}

func (b *participantListCreator) Len() int {
	return len(b.participants)
}

func (b *participantListCreator) Build() (map[int64]*Participant, error) {

	result := make(map[int64]*Participant)

	if b.authorID == 0 {
		return nil, ErrMissingArgument
	}

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

		if isFriend, err := b.eventManager.parent.Friends.IsFriend(participant.id, b.authorID); err != nil {
			return nil, err
		} else if isFriend {
			result[participant.id] = participant
		} else {
			log.Printf("* WARNING: CREATE PARTICIPANT LIST -> USER %v TRIED TO ADD USER %v BUT THEY ARE NOT FRIENDS\n",
				b.authorID, participant.id)
		}
	}

	if len(result) == 0 {
		return nil, ErrParticipantsRequired
	}

	return result, nil
}
