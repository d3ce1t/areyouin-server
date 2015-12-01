package main

import (
	proto "areyouin/protocol"
	"sort"
)

func NewInbox(user_id uint64) *Inbox {
	return &Inbox{user_id: user_id}
}

type Inbox struct {
	user_id uint64
	events  []*proto.Event
}

func (in *Inbox) Add(event *proto.Event) {
	in.events = append(in.events, event)
}

func (in *Inbox) GetAll() []*proto.Event {
	return in.events
}

func (in *Inbox) Len() int {
	return len(in.events)
}

func (in Inbox) RemoveOldEvents() {

}

type ByDate []*proto.Event

func (a ByDate) Len() int           { return len(a) }
func (a ByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool { return a[i].StartDate < a[j].StartDate }

// Delegate sorting to the client.
func (in *Inbox) Sort() {
	sort.Sort(ByDate(in.events))
}
