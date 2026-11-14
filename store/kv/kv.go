package kv

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/Sid-80/Redis-go/store"
)

type Store struct {
	data    map[string]string
	maxSize int //0 = unlimited
	mu      sync.RWMutex
}

var ErrMaxSize = errors.New("You have reached max limit to store keys!!")

// constructor function
func NewStore(maxSize int) *Store {
	return &Store{
		data:    make(map[string]string),
		maxSize: maxSize,
	}
}

func (s *Store) Clone() *Store {
	cp := &Store{
		data:    make(map[string]string, len(s.data)),
		maxSize: s.maxSize,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range s.data {
		cp.data[k] = v
	}

	return cp
}

// this function means attaching something to struct i.e. Store.Get("key1")
func (s *Store) Get(Key string) (string, error) {
	if Key == "" {
		return "", store.ErrEmptyKey
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

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
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.data[Key]

	if s.maxSize > 0 && len(s.data) >= s.maxSize && (!err) {
		return ErrMaxSize
	}

	s.data[Key] = Value
	return nil
}

func (s *Store) Del(Key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, Key)
	return
}

func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))

	for key, _ := range s.data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func (s *Store) Rename(oldKey string, newKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, _ := s.data[oldKey]

	if val != "" {
		delete(s.data, oldKey)
		s.data[newKey] = val
	}
}

func (s *Store) Pop(key string) (string, bool) {
	keys := s.Keys()

	if len(keys) > 0 {
		return "", false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lastKey := keys[len(keys)]

	val, _ := s.data[lastKey]

	delete(s.data, lastKey)

	return val, true
}

func (s *Store) INCR(Key string) (int, error) {
	if Key == "" {
		return 0, store.ErrEmptyKey
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	current := 0
	if val, ok := s.data[Key]; ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			return 0, fmt.Errorf("INCR Error: %q", err)
		}
		current = v
	}

	current++
	time.Sleep(time.Microsecond)

	s.data[Key] = strconv.Itoa(current)

	return current, nil
}
