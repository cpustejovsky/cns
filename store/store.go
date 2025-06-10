package store

import (
	"fmt"
	"sync"
)

type Store struct {
	sync.RWMutex
	m map[string]string
}

func New() Store {
	return Store{
		RWMutex: sync.RWMutex{},
		m:       make(map[string]string),
	}
}

type ErrorNoSuchKey struct {
	key string
}

func (e ErrorNoSuchKey) Error() string {
	return fmt.Sprintf("no value for key '%s'", e.key)
}

func (s *Store) Put(key, value string) error {
	s.Lock()
	defer s.Unlock()
	s.m[key] = value
	return nil
}

func (s *Store) Get(key string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	val, ok := s.m[key]
	if !ok {
		return val, ErrorNoSuchKey{key: key}
	}
	return val, nil
}

func (s *Store) Delete(key string) error {
	s.Lock()
	defer s.Unlock()
	delete(s.m, key)
	return nil
}
