package main

import "sync"

type Store struct {
	mu     sync.RWMutex
	values map[string]string
}

func NewStore() *Store {
	return &Store{values: make(map[string]string)}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.values[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.values[key]
	return value, ok
}
