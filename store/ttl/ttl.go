package ttl

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/Sid-80/Redis-go/store"
)

// Compile-time check that *TTLStore implements store.Storer.
var _ store.Storer = (*TTLStore)(nil)

// TTLStore is a key/value store where every key expires ttl after it was set.
// Expired keys are removed lazily (when they are touched), not by a background job.
//
// Methods that hold the lock must touch t.data directly and never call
// another locking method, because sync.Mutex is not re-entrant.
type TTLStore struct {
	// A plain Mutex, not RWMutex: even Get may delete an expired key, so every method writes.
	mu   sync.Mutex
	data map[string]ttlEntry
	ttl  time.Duration
}

type ttlEntry struct {
	value     string
	expiresAt time.Time
}

func (e ttlEntry) expired(now time.Time) bool {
	return now.After(e.expiresAt)
}

func NewTTLStore(ttl time.Duration) *TTLStore {
	return &TTLStore{
		data: make(map[string]ttlEntry),
		ttl:  ttl,
	}
}

// liveEntry returns the entry for key, deleting it if it has expired.
// It must be called with the lock held.
func (t *TTLStore) liveEntry(key string, now time.Time) (ttlEntry, bool) {
	entry, ok := t.data[key]
	if !ok {
		return ttlEntry{}, false
	}
	if entry.expired(now) {
		delete(t.data, key)
		return ttlEntry{}, false
	}
	return entry, true
}

// evictExpired removes every expired key. It must be called with the lock held.
func (t *TTLStore) evictExpired(now time.Time) {
	for k, e := range t.data {
		if e.expired(now) {
			delete(t.data, k)
		}
	}
}

func (t *TTLStore) Set(key string, value string) error {
	if key == "" {
		return store.ErrEmptyKey
	}

	t.mu.Lock()
	defer t.mu.Unlock()

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

	t.mu.Lock()
	defer t.mu.Unlock()

	entry, ok := t.liveEntry(key, time.Now())
	if !ok {
		return "", store.ErrKeyDoesNotExist
	}
	return entry.value, nil
}

func (t *TTLStore) Keys() []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.evictExpired(time.Now())

	keys := make([]string, 0, len(t.data))
	for k := range t.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (t *TTLStore) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.evictExpired(time.Now())
	return len(t.data)
}

func (t *TTLStore) Del(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.data, key)
}

// Rename moves oldKey to newKey, keeping its original expiry time.
// It does nothing if oldKey is missing or expired.
func (t *TTLStore) Rename(oldKey string, newKey string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	entry, ok := t.liveEntry(oldKey, time.Now())
	if !ok {
		return
	}
	delete(t.data, oldKey)
	t.data[newKey] = entry
}

// Pop removes and returns the value of the alphabetically last live key.
// It returns false if there are no live keys.
func (t *TTLStore) Pop() (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.evictExpired(time.Now())
	if len(t.data) == 0 {
		return "", false
	}

	lastKey := ""
	for k := range t.data {
		if k > lastKey {
			lastKey = k
		}
	}

	val := t.data[lastKey].value
	delete(t.data, lastKey)
	return val, true
}

// INCR adds 1 to the integer stored at key and returns the new value.
// An existing key keeps its expiry time (like Redis); a missing or
// expired key starts at 0 and gets a fresh TTL.
func (t *TTLStore) INCR(key string) (int, error) {
	if key == "" {
		return 0, store.ErrEmptyKey
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	entry, ok := t.liveEntry(key, now)

	current := 0
	if ok {
		v, err := strconv.Atoi(entry.value)
		if err != nil {
			return 0, fmt.Errorf("INCR %q: value is not an integer: %w", key, err)
		}
		current = v
	} else {
		entry.expiresAt = now.Add(t.ttl)
	}

	current++
	entry.value = strconv.Itoa(current)
	t.data[key] = entry
	return current, nil
}
