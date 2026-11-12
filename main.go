package main

import (
	"encoding/base64"
	"fmt"

	"github.com/Sid-80/Redis-go/store"
	"github.com/Sid-80/Redis-go/store/kv"
)

func main() {
	plain := kv.NewStore(20)
	logger := NewLoggingMiddleware(plain)
	metrics := NewMetricMiddleware(logger)

	encrypted, err := SetKeyWithEncryptionStore(metrics, "a", "Its all set without ttl")

	if err != nil {
		fmt.Println((err))
		return
	}

	fmt.Println(encrypted)

	metrics.Report()
}

func SetKeyWithEncryptionStore(s store.Storer, key, val string) (string, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(val))

	if err := s.Set(key, encoded); err != nil {
		fmt.Println("Error")
		return "", err
	}

	return s.Get(key)
}
