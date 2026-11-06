package main

import (
	"errors"
	"sort"
)

type Store struct {
	data map[string]string
}

// this function means attaching something to struct i.e. Store.Get("key1")
func (s *Store) Get(Key string) (string, error) {
	val, exists := s.data[Key]

	if !exists {
		return "", errors.New("Key doesn't exist!!")
	}

	return val, nil
}

func (s *Store) Set(Key string, Value string) {
	s.data[Key] = Value
	return
}

func (s *Store) Del(Key string) {
	delete(s.data, Key)
	return
}

func (s *Store) Keys() []string {
	keys := make([]string, 0, len(s.data))

	for key, _ := range s.data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func (s *Store) len() int {
	return len(s.data)
}

func (s *Store) Rename(oldKey string, newKey string) {
	val, _ := s.Get(oldKey)

	if val != "" {
		s.Del(oldKey)
		s.Set(newKey, val)
	}
}

func (s *Store) Pop(key string) (string, bool) {
	keys := s.Keys()

	if len(keys) > 0 {
		return "", false
	}

	lastKey := keys[len(keys)]

	val, _ := s.Get(lastKey)

	s.Del(lastKey)

	return val, true
}

// constructor function
func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}
