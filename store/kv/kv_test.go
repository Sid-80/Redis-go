package kv

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/Sid-80/Redis-go/store"
)

// mustSet fails the test immediately if Set returns an error.
func mustSet(t *testing.T, s *Store, key, value string) {
	t.Helper()
	if err := s.Set(key, value); err != nil {
		t.Fatalf("Set(%q, %q): %v", key, value, err)
	}
}

func TestKeys_ReturnsAllKeysSorted(t *testing.T) {
	s := NewStore(3)
	mustSet(t, s, "charlie", "3")
	mustSet(t, s, "alpha", "1")
	mustSet(t, s, "bravo", "2")

	got := s.Keys()
	want := []string{"alpha", "bravo", "charlie"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
}

func TestKeys_EmptyStore(t *testing.T) {
	s := NewStore(3)
	if got := s.Keys(); len(got) != 0 {
		t.Errorf("Keys() on empty store returned %d keys, want 0", len(got))
	}
}

func TestGetSet_EmptyKey(t *testing.T) {
	s := NewStore(3)

	if _, err := s.Get(""); !errors.Is(err, store.ErrEmptyKey) {
		t.Errorf("Get(\"\") error = %v, want %v", err, store.ErrEmptyKey)
	}
	if err := s.Set("", "v"); !errors.Is(err, store.ErrEmptyKey) {
		t.Errorf("Set(\"\") error = %v, want %v", err, store.ErrEmptyKey)
	}
}

func TestGetSet_RoundTrip(t *testing.T) {
	s := NewStore(3)
	mustSet(t, s, "Hello", "World")

	val, err := s.Get("Hello")
	if err != nil || val != "World" {
		t.Errorf("Get(\"Hello\") = %q, %v; want \"World\", nil", val, err)
	}

	if _, err := s.Get("missing"); !errors.Is(err, store.ErrKeyDoesNotExist) {
		t.Errorf("Get(\"missing\") error = %v, want %v", err, store.ErrKeyDoesNotExist)
	}
}

func TestSet_MaxSize(t *testing.T) {
	s := NewStore(2)
	mustSet(t, s, "a", "1")
	mustSet(t, s, "b", "2")

	if err := s.Set("c", "3"); !errors.Is(err, store.ErrStoreFull) {
		t.Errorf("Set on full store error = %v, want %v", err, store.ErrStoreFull)
	}

	// Overwriting an existing key must still work when the store is full.
	if err := s.Set("a", "updated"); err != nil {
		t.Errorf("overwriting existing key on full store: %v", err)
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name      string
		setup     map[string]string
		deleteKey string
		wantLen   int
	}{
		{"deletes existing key", map[string]string{"a": "1", "b": "2"}, "a", 1},
		{"no-op on missing key", map[string]string{"a": "1", "b": "2"}, "x", 2},
		{"no-op on empty store", map[string]string{}, "x", 0},
		{"deletes last remaining key", map[string]string{"x": "1"}, "x", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewStore(3)
			for k, v := range tc.setup {
				mustSet(t, s, k, v)
			}

			s.Del(tc.deleteKey)

			if got := s.Len(); got != tc.wantLen {
				t.Errorf("after Del(%q): Len() = %d, want %d", tc.deleteKey, got, tc.wantLen)
			}
			if _, err := s.Get(tc.deleteKey); err == nil {
				t.Errorf("after Del(%q): key is still present", tc.deleteKey)
			}
		})
	}
}

func TestRename(t *testing.T) {
	s := NewStore(0)
	mustSet(t, s, "old", "v")
	mustSet(t, s, "empty", "")

	s.Rename("old", "new")
	if _, err := s.Get("old"); err == nil {
		t.Error("old key still exists after Rename")
	}
	if got, _ := s.Get("new"); got != "v" {
		t.Errorf("Get(\"new\") = %q, want \"v\"", got)
	}

	// A key whose value is "" must still be renamed.
	s.Rename("empty", "empty2")
	if _, err := s.Get("empty2"); err != nil {
		t.Errorf("key with empty value was not renamed: %v", err)
	}

	// Renaming a missing key is a no-op.
	s.Rename("nope", "whatever")
	if _, err := s.Get("whatever"); err == nil {
		t.Error("Rename of missing key created a new key")
	}
}

func TestPop(t *testing.T) {
	s := NewStore(0)

	if _, ok := s.Pop(); ok {
		t.Fatal("Pop() on empty store returned ok = true")
	}

	mustSet(t, s, "a", "1")
	mustSet(t, s, "c", "3")
	mustSet(t, s, "b", "2")

	val, ok := s.Pop()
	if !ok || val != "3" {
		t.Errorf("Pop() = %q, %v; want \"3\", true", val, ok)
	}
	if s.Len() != 2 {
		t.Errorf("Len() after Pop = %d, want 2", s.Len())
	}
}

func TestINCR(t *testing.T) {
	s := NewStore(0)

	if got, err := s.INCR("n"); err != nil || got != 1 {
		t.Errorf("INCR on missing key = %d, %v; want 1, nil", got, err)
	}
	if got, err := s.INCR("n"); err != nil || got != 2 {
		t.Errorf("second INCR = %d, %v; want 2, nil", got, err)
	}

	mustSet(t, s, "word", "hello")
	if _, err := s.INCR("word"); err == nil {
		t.Error("INCR on non-integer value returned nil error")
	}

	if _, err := s.INCR(""); !errors.Is(err, store.ErrEmptyKey) {
		t.Errorf("INCR(\"\") error = %v, want %v", err, store.ErrEmptyKey)
	}
}

func TestINCR_Concurrent(t *testing.T) {
	const workers = 1000
	s := NewStore(0)

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.INCR("n"); err != nil {
				t.Errorf("INCR: %v", err)
			}
		}()
	}
	wg.Wait()

	if got, _ := s.Get("n"); got != "1000" {
		t.Fatalf("after %d concurrent INCRs got %s, want 1000", workers, got)
	}
}
