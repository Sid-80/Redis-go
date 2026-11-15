package kv

import (
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/Sid-80/Redis-go/store"
)

// Compile-time check that *Store implements store.Storer.
var _ store.Storer = (*Store)(nil)

// Store is an in-memory key/value store that is safe for concurrent use.
//
// Methods that hold the lock must touch s.data directly and never call
// another locking method, because sync.RWMutex is not re-entrant.
type Store struct {
	mu      sync.RWMutex
	data    map[string]string
	maxSize int // 0 = unlimited
}

func NewStore(maxSize int) *Store {
	return &Store{
		data:    make(map[string]string),
		maxSize: maxSize,
	}
}

func (s *Store) Clone() *Store {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cp := &Store{
		data:    make(map[string]string, len(s.data)),
		maxSize: s.maxSize,
	}
	for k, v := range s.data {
		cp.data[k] = v
	}
	return cp
}

func (s *Store) Get(key string) (string, error) {
	if key == "" {
		return "", store.ErrEmptyKey
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	val, exists := s.data[key]
	if !exists {
		return "", store.ErrKeyDoesNotExist
	}
	return val, nil
}

func (s *Store) Set(key string, value string) error {
	if key == "" {
		return store.ErrEmptyKey
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Overwriting an existing key is always allowed; only new keys count towards maxSize.
	_, exists := s.data[key]
	if s.maxSize > 0 && len(s.data) >= s.maxSize && !exists {
		return store.ErrStoreFull
	}

	s.data[key] = value
	return nil
}

func (s *Store) Del(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return sortedKeys(s.data)
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// Rename moves the value at oldKey to newKey. It does nothing if oldKey is missing.
func (s *Store) Rename(oldKey string, newKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.data[oldKey]
	if !ok {
		return
	}
	delete(s.data, oldKey)
	s.data[newKey] = val
}

// Pop removes and returns the value of the alphabetically last key.
// It returns false if the store is empty.
func (s *Store) Pop() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.data) == 0 {
		return "", false
	}

	keys := sortedKeys(s.data)
	lastKey := keys[len(keys)-1]

	val := s.data[lastKey]
	delete(s.data, lastKey)
	return val, true
}

// INCR adds 1 to the integer stored at key (a missing key starts at 0)
// and returns the new value.
func (s *Store) INCR(key string) (int, error) {
	if key == "" {
		return 0, store.ErrEmptyKey
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	current := 0
	if val, ok := s.data[key]; ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			return 0, fmt.Errorf("INCR %q: value is not an integer: %w", key, err)
		}
		current = v
	} else if s.maxSize > 0 && len(s.data) >= s.maxSize {
		// INCR on a missing key creates it, so it must respect maxSize like Set.
		return 0, store.ErrStoreFull
	}

	current++
	s.data[key] = strconv.Itoa(current)
	return current, nil
}

// sortedKeys must be called with the lock held.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
