package main

import (
	"sync"
)

func NewSessionsMap() *SessionsMap {
	object := &SessionsMap{
		m: make(map[int64]*AyiSession),
	}
	return object
}

type SessionsMap struct {
	mutex sync.RWMutex
	m     map[int64]*AyiSession
}

func (sm *SessionsMap) Get(key int64) (v *AyiSession, ok bool) {
	defer sm.mutex.RUnlock()
	sm.mutex.RLock()
	v, ok = sm.m[key]
	return
}

func (sm *SessionsMap) Put(key int64, session *AyiSession) {
	defer sm.mutex.Unlock()
	sm.mutex.Lock()
	sm.m[key] = session
}

func (sm *SessionsMap) Remove(key int64) {
	defer sm.mutex.Unlock()
	sm.mutex.Lock()
	delete(sm.m, key)
}

func (sm *SessionsMap) Len() int {
	defer sm.mutex.RUnlock()
	sm.mutex.RLock()
	return len(sm.m)
}

func (sm *SessionsMap) Keys() (keys []int64) {
	defer sm.mutex.RUnlock()
	sm.mutex.RLock()
	keys = make([]int64, 0, len(sm.m))
	for k, _ := range sm.m {
		keys = append(keys, k)
	}
	return keys
}
