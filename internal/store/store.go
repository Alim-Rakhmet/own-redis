package store

import (
	"sync"
	"time"
)

type item struct {
	value   string
	timeOut int64
}

type Store struct {
	mutex sync.RWMutex
	store map[string]item
}

func NewStore() *Store {
	return &Store{
		store: make(map[string]item),
	}
}

func (s *Store) Set(key, value string, px int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var timeOut int64
	if px > 0 {
		timeOut = time.Now().UnixMilli() + px
	}

	s.store[key] = item{
		value:   value,
		timeOut: timeOut,
	}
}

func (s *Store) Get(key string) (string, bool) {
	s.mutex.RLock()
	item, exists := s.store[key]
	s.mutex.RUnlock()

	if !exists {
		return "", false
	}

	if item.timeOut > 0 && item.timeOut < time.Now().UnixMilli() {
		s.mutex.Lock()

		if it, ok := s.store[key]; ok && item.timeOut == it.timeOut {
			delete(s.store, key)
		}
		s.mutex.Unlock()

		return "", false
	}

	return item.value, true
}
