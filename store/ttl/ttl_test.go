package ttl

import (
	"fmt"
	"testing"
	"time"
)

func TestTTLStore_GetBeforeExpiry(t *testing.T) {
	s := NewTTLStore(1 * time.Second)

	err1 := s.Set("k", "v")

	if err1 != nil {
		fmt.Println(err1)
	}

	val, err2 := s.Get("k")

	if err2 != nil || val != "v" {
		t.Fatalf("Expected v, got %q %v", val, err2)
	}

}
func TestTTLStore_GetAfterExpiry(t *testing.T) {
	s := NewTTLStore(50 * time.Millisecond)

	err1 := s.Set("k", "v")

	if err1 != nil {
		fmt.Println(err1)
	}

	time.Sleep(100 * time.Millisecond)

	_, err2 := s.Get("k")

	if err2 == nil {
		t.Fatalf("Expected error after ttl, got nil")
	}

}

func TestTTLStore_ExpiredKeyEvicted(t *testing.T) {
	s := NewTTLStore(50 * time.Millisecond)

	err1 := s.Set("k", "v")

	if err1 != nil {
		fmt.Println(err1)
	}

	time.Sleep(100 * time.Millisecond)

	_, err2 := s.Get("k")

	if err2 == nil || s.Len() != 0 {
		t.Fatalf("Expected key to be evicted!!")
	}
}
