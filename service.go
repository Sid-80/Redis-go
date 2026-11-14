package main

import (
	"encoding/base64"
	"fmt"

	"github.com/Sid-80/Redis-go/store"
)

func SetKeyWithEncryptionStore(s store.Storer, key, val string) (string, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(val))

	if err := s.Set(key, encoded); err != nil {
		fmt.Println("Error")
		return "", err
	}

	return s.Get(key)
}
