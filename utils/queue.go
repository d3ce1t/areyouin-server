package utils

import (
	"sync"
)

type Queue struct {
	lhm *LinkedHashMap
	m   sync.RWMutex
}

func NewQueue() *Queue {
	return &Queue{
		lhm: NewLinkedHashMap(),
	}
}

func (q *Queue) Add(item interface{}) {
	defer q.m.Unlock()
	q.m.Lock()
	q.lhm.PushBack(item)
}

func (q *Queue) AddWithKey(key string, item interface{}) {
	defer q.m.Unlock()
	q.m.Lock()
	if key != "" {
		q.lhm.PushBackWithCollapseKey(key, item)
	} else {
		q.lhm.PushBack(item)
	}
}

func (q *Queue) Element() interface{} {
	defer q.m.RUnlock()
	q.m.RLock()
	item := q.lhm.Front()
	if item != nil {
		return item.Value
	}
	return nil
}

func (q *Queue) Remove() interface{} {
	defer q.m.Unlock()
	q.m.Lock()
	item := q.lhm.Front()
	if item != nil {
		return q.lhm.Remove(item)
	}
	return nil
}
