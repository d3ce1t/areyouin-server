package model

import "sync"

type UserEvents struct {
	mutex  sync.RWMutex
	events map[int64]map[int64]bool // userID -> eventID
}

func newUserEvents() *UserEvents {
	object := &UserEvents{
		events: make(map[int64]map[int64]bool),
	}
	return object
}

func (u *UserEvents) Insert(userID int64, eventID int64) {

	defer u.mutex.Unlock()
	u.mutex.Lock()

	if _, exist := u.events[userID]; !exist {
		u.events[userID] = make(map[int64]bool)
	}

	u.events[userID][eventID] = true

	/*listEntry := u.eventsByUser[userID]
	i := sort.Search(len(listEntry), func(i int) bool {
		return listEntry[i].EndDate.After(entry.EndDate) || listEntry[i].EndDate.Equal(entry.EndDate)
	})

	if i < len(listEntry) && listEntry[i].EndDate.Equal(entry.EndDate) {
		// entry is present at listEntry[i]
		if listEntry[i].EventID != entry.EventID {
			listEntry = append(listEntry, nil)
			copy(listEntry[i+1:], listEntry[i:])
			listEntry[i] = listEntry[i+1]
			listEntry[i+1] = entry
			u.eventsByUser[userID] = listEntry
		}

	} else {
		// entry is not present in entryList,
		// but i is the index where it would be inserted.
		listEntry = append(listEntry, nil)
		copy(listEntry[i+1:], listEntry[i:])
		listEntry[i] = entry
		u.eventsByUser[userID] = listEntry
	}*/
}

func (u *UserEvents) Remove(userID int64, eventID int64) {
	defer u.mutex.Unlock()
	u.mutex.Lock()
	delete(u.events[userID], eventID)
	if len(u.events[userID]) == 0 {
		delete(u.events, userID)
	}
}

func (u *UserEvents) FindAll(userID int64) []int64 {
	defer u.mutex.RUnlock()
	u.mutex.RLock()

	eventIDs := make([]int64, 0, len(u.events[userID]))

	for k := range u.events[userID] {
		eventIDs = append(eventIDs, k)
	}

	return eventIDs
}

func (u *UserEvents) Len() int {
	defer u.mutex.RUnlock()
	u.mutex.RLock()
	return len(u.events)
}
