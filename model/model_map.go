package model

import (
	"sync"
)

func newModelsMap() *ModelsMap {
	object := &ModelsMap{
		m: make(map[string]*AyiModel),
	}
	return object
}

type ModelsMap struct {
	mutex sync.RWMutex
	m     map[string]*AyiModel
}

func (sm *ModelsMap) Get(key string) (v *AyiModel, ok bool) {
	defer sm.mutex.RUnlock()
	sm.mutex.RLock()
	v, ok = sm.m[key]
	return
}

func (sm *ModelsMap) Put(key string, model *AyiModel) {
	defer sm.mutex.Unlock()
	sm.mutex.Lock()
	sm.m[key] = model
}

func (sm *ModelsMap) Remove(key string) {
	defer sm.mutex.Unlock()
	sm.mutex.Lock()
	delete(sm.m, key)
}

func (sm *ModelsMap) Len() int {
	defer sm.mutex.RUnlock()
	sm.mutex.RLock()
	return len(sm.m)
}

func (sm *ModelsMap) Keys() (keys []string) {
	defer sm.mutex.RUnlock()
	sm.mutex.RLock()
	keys = make([]string, 0, len(sm.m))
	for k, _ := range sm.m {
		keys = append(keys, k)
	}
	return keys
}
