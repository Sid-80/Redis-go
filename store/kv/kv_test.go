package kv

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/Sid-80/Redis-go/store"
)

func TestKeys_ReturnsAllKeysSorted(t *testing.T) {
	store := NewStore(3)
	store.Set("charlie", "3")
	store.Set("alpha", "1")
	store.Set("bravo", "2")

	got := store.Keys()

	want := []string{"alpha", "bravo", "charlie"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Opps!! Keys are different!! \n Got: %v \n Want: %v", got, want)
	}
}

func TestGetSet_EmptyKeys(t *testing.T) {
	store2 := NewStore(3)

	if _, err := store2.Get(""); err == nil || !errors.Is(err, store.ErrEmptyKey) {
		t.Error("Get() failed!!")
	}

}

func TestKeys_EmptyStore(t *testing.T) {
	store := NewStore(3)
	got := store.Keys()
	if len(got) != 0 {
		t.Errorf("Opps!! Its length is not 0 !! \n len of Got: %v", len(got))
	}
}

func TestGetSet_RoundTrip(t *testing.T) {
	store2 := NewStore(3)
	store2.Set("Hello", "World")

	val, err := store2.Get("Hello")

	if val != "World" || err != nil {
		t.Errorf("Invalid value while Get and Set \n Value: %v", val)
	}

	_, err2 := store2.Get("h2")

	fmt.Println(err2)

	if err2 != store.ErrKeyDoesNotExist {
		t.Error("Issue in Get function its returing true for value which is not Set")
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
			store := NewStore(3)
			for k, v := range tc.setup {
				store.Set(k, v)
			}

			store.Del(tc.deleteKey)

			if got := store.Len(); got != tc.wantLen {
				t.Errorf("after Delete (%q): len(): %d, want: %d", tc.deleteKey, got, tc.wantLen)
			}

			if _, err := store.Get(tc.deleteKey); err == nil {
				t.Errorf("after Delete (%q) is still present", tc.deleteKey)
			}

		})
	}

}
