package ttl

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Sid-80/Redis-go/store"
)

// mustSet fails the test immediately if Set returns an error.
func mustSet(t *testing.T, s *TTLStore, key, value string) {
	t.Helper()
	if err := s.Set(key, value); err != nil {
		t.Fatalf("Set(%q, %q): %v", key, value, err)
	}
}

func TestTTLStore_GetBeforeExpiry(t *testing.T) {
	s := NewTTLStore(time.Second)
	mustSet(t, s, "k", "v")

	val, err := s.Get("k")
	if err != nil || val != "v" {
		t.Fatalf("Get(\"k\") = %q, %v; want \"v\", nil", val, err)
	}
}

func TestTTLStore_GetAfterExpiry(t *testing.T) {
	s := NewTTLStore(50 * time.Millisecond)
	mustSet(t, s, "k", "v")

	time.Sleep(100 * time.Millisecond)

	if _, err := s.Get("k"); !errors.Is(err, store.ErrKeyDoesNotExist) {
		t.Fatalf("Get after expiry error = %v, want %v", err, store.ErrKeyDoesNotExist)
	}
}

func TestTTLStore_ExpiredKeyEvicted(t *testing.T) {
	s := NewTTLStore(50 * time.Millisecond)
	mustSet(t, s, "k", "v")

	time.Sleep(100 * time.Millisecond)

	if _, err := s.Get("k"); err == nil {
		t.Fatal("expected error for expired key")
	}
	if got := s.Len(); got != 0 {
		t.Fatalf("Len() after expiry = %d, want 0", got)
	}
}

func TestTTLStore_KeysAndLenSkipExpired(t *testing.T) {
	s := NewTTLStore(50 * time.Millisecond)
	mustSet(t, s, "a", "1")
	mustSet(t, s, "b", "2")

	time.Sleep(100 * time.Millisecond)

	// Never read with Get, but Keys and Len must still not report them.
	if got := s.Keys(); len(got) != 0 {
		t.Errorf("Keys() after expiry = %v, want []", got)
	}
	if got := s.Len(); got != 0 {
		t.Errorf("Len() after expiry = %d, want 0", got)
	}
}

func TestTTLStore_RenameKeepsExpiry(t *testing.T) {
	s := NewTTLStore(80 * time.Millisecond)
	mustSet(t, s, "old", "v")

	time.Sleep(50 * time.Millisecond)
	s.Rename("old", "new")

	if got, err := s.Get("new"); err != nil || got != "v" {
		t.Fatalf("Get(\"new\") = %q, %v; want \"v\", nil", got, err)
	}

	// The renamed key keeps the original deadline, so it is gone soon after.
	time.Sleep(50 * time.Millisecond)
	if _, err := s.Get("new"); err == nil {
		t.Error("renamed key got a fresh TTL; want original expiry kept")
	}
}

func TestTTLStore_Pop(t *testing.T) {
	s := NewTTLStore(time.Second)

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
}

func TestTTLStore_INCR(t *testing.T) {
	s := NewTTLStore(time.Second)

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
}

func TestTTLStore_INCR_Concurrent(t *testing.T) {
	const workers = 1000
	s := NewTTLStore(time.Minute)

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
