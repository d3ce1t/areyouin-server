package model

type ParticipantList struct {
	numAttendees int
	numGuests    int
	participants map[int64]*Participant
}

func newParticipantList() *ParticipantList {
	return &ParticipantList{
		participants: make(map[int64]*Participant),
	}
}

func (l *ParticipantList) Get(id int64) (*Participant, bool) {
	v, ok := l.participants[id]
	return v, ok
}

func (l *ParticipantList) Ids() []int64 {
	keys := make([]int64, 0, len(l.participants))
	for k := range l.participants {
		keys = append(keys, k)
	}
	return keys
}

func (l *ParticipantList) AsSlice() []*Participant {
	values := make([]*Participant, 0, len(l.participants))
	for _, v := range l.participants {
		values = append(values, v)
	}
	return values
}

func (l *ParticipantList) NumAttendees() int {
	return l.numAttendees
}

func (l *ParticipantList) NumGuests() int {
	return l.numGuests
}

func (l *ParticipantList) Equal(other *ParticipantList) bool {

	if len(l.participants) != len(other.participants) ||
		l.numGuests != other.numGuests || l.numAttendees != other.numAttendees {
		return false
	}

	for k, p1 := range l.participants {
		if p2, ok := other.participants[k]; !ok || !p1.Equal(p2) {
			return false
		}
	}

	return true
}

func (l *ParticipantList) Clone() *ParticipantList {
	copy := new(ParticipantList)
	*copy = *l
	copy.participants = make(map[int64]*Participant)
	for _, p := range l.participants {
		copy.participants[p.id] = p.Clone()
	}
	return copy
}
