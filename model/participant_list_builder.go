package model

import (
	"log"
	"peeple/areyouin/api"
	"time"
)

type ParticipantAdder interface {
	AddUserAccount(u *UserAccount) ParticipantAdder
	AddFriend(f *Friend) ParticipantAdder
	AddParticipant(p *Participant) ParticipantAdder
	AddUserID(UID int64) ParticipantAdder
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

func (b *participantListCreator) AddUserAccount(u *UserAccount) ParticipantAdder {
	b.participants[u.id] = u.AsParticipant()
	return b
}

func (b *participantListCreator) AddFriend(f *Friend) ParticipantAdder {
	b.participants[f.id] = NewParticipant(f.id, f.name, api.AttendanceResponse_NO_RESPONSE,
		api.InvitationStatus_SERVER_DELIVERED)
	return b
}

func (b *participantListCreator) AddParticipant(p *Participant) ParticipantAdder {
	b.participants[p.id] = p.Clone()
	return b
}

func (b *participantListCreator) AddUserID(UID int64) ParticipantAdder {
	b.participants[UID] = UID
	return b
}

func (b *participantListCreator) Len() int {
	return len(b.participants)
}

func (b *participantListCreator) Build() (*ParticipantList, error) {

	if b.ownerID == 0 {
		return nil, ErrMissingArgument
	}

	list := newParticipantList()

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
		participant.nameTS = b.timestamp
		participant.responseTS = b.timestamp
		participant.statusTS = b.timestamp
		list.participants[participant.id] = participant

		if participant.response == api.AttendanceResponse_ASSIST {
			list.numAttendees++
		}
	}

	list.numGuests = len(list.participants)

	if len(list.participants) == 0 {
		return nil, ErrParticipantsRequired
	}

	return list, nil
}
