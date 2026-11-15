package store

import "errors"

var (
	ErrEmptyKey        = errors.New("key is empty")
	ErrStoreFull       = errors.New("store is full")
	ErrKeyDoesNotExist = errors.New("key does not exist")
)

// Storer is the behaviour every store (and every middleware wrapping a store) must provide.
type Storer interface {
	Get(key string) (string, error)
	Set(key string, value string) error
	Keys() []string
	Del(key string)
	Len() int
	INCR(key string) (int, error)
}
