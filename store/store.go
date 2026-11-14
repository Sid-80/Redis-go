package store

import "errors"

var ErrEmptyKey = errors.New("Key is empty!!")
var ErrStoreFull = errors.New("Store is full!!")
var ErrKeyDoesNotExist = errors.New("Key doesn't exist!!")

type Storer interface {
	Get(Key string) (string, error)
	Set(Key string, Value string) error
	Keys() []string
	Del(Key string)
	Len() int
	INCR(Key string) (int, error)
}
