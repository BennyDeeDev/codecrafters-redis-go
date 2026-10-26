package main

import (
	"sync"
	"time"
)

type Entry struct {
	value     string
	expiresAt time.Time
}

type Store struct {
	mu     sync.RWMutex
	values map[string]Entry
}

func NewStore() *Store {
	return &Store{values: make(map[string]Entry)}
}

func (s *Store) Set(key string, entry Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.values[key] = entry
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.values[key]
	if !ok {
		return "", false
	}
	if !entry.expiresAt.IsZero() && !time.Now().Before(entry.expiresAt) {
		delete(s.values, key)
		return "", false
	}

	return entry.value, true
}
