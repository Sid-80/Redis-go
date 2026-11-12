package ttl

import (
	"sort"
	"time"

	"github.com/Sid-80/Redis-go/store"
)

type TTLStore struct {
	data map[string]ttlEntry
	ttl  time.Duration
}

type ttlEntry struct {
	value     string
	expiresAt time.Time
}

func NewTTLStore(ttl time.Duration) *TTLStore {
	return &TTLStore{
		data: make(map[string]ttlEntry),
		ttl:  ttl,
	}
}

func (t *TTLStore) Keys() []string {
	keys := make([]string, 0, t.Len())

	for key, _ := range t.data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func (t *TTLStore) Set(key string, value string) error {

	if key == "" {
		return store.ErrEmptyKey
	}

	t.data[key] = ttlEntry{
		value:     value,
		expiresAt: time.Now().Add(t.ttl),
	}

	return nil
}

func (t *TTLStore) Get(key string) (string, error) {
	if key == "" {
		return "", store.ErrEmptyKey
	}

	entry, ok := t.data[key]

	if !ok || time.Now().After(entry.expiresAt) {
		// Lazy eviction - no background job scanning for expired keys, deletion on demand
		delete(t.data, key)

		return "", store.ErrKeyDoesNotExist
	}

	return entry.value, nil
}

func (t *TTLStore) Len() int {
	return len(t.data)
}

func (t *TTLStore) Del(Key string) {
	delete(t.data, Key)
	return
}

func (t *TTLStore) Rename(oldKey string, newKey string) {
	val, _ := t.Get(oldKey)

	if val != "" {
		t.Del(oldKey)
		t.Set(newKey, val)
	}
}

func (s *TTLStore) Pop(key string) (string, bool) {
	keys := s.Keys()

	if len(keys) > 0 {
		return "", false
	}

	lastKey := keys[len(keys)]

	val, _ := s.Get(lastKey)

	s.Del(lastKey)

	return val, true
}
