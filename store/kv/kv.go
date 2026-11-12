package kv

import (
	"errors"
	"sort"

	"github.com/Sid-80/Redis-go/store"
)

type Store struct {
	data    map[string]string
	maxSize int //0 = unlimited
}

var ErrMaxSize = errors.New("You have reached max limit to store keys!!")

// constructor function
func NewStore(maxSize int) *Store {
	return &Store{
		data:    make(map[string]string),
		maxSize: maxSize,
	}
}

// this function means attaching something to struct i.e. Store.Get("key1")
func (s *Store) Get(Key string) (string, error) {
	if Key == "" {
		return "", store.ErrEmptyKey
	}

	val, exists := s.data[Key]

	if !exists {
		return "", store.ErrKeyDoesNotExist
	}

	return val, nil
}

func (s *Store) Set(Key string, Value string) error {
	if Key == "" {
		return store.ErrEmptyKey
	}

	_, err := s.Get(Key)

	if s.maxSize > 0 && s.Len() >= s.maxSize && (err == store.ErrKeyDoesNotExist) {
		return ErrMaxSize
	}

	s.data[Key] = Value
	return nil
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

func (s *Store) Len() int {
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
