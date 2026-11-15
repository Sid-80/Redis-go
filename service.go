package main

import (
	"encoding/base64"
	"fmt"

	"github.com/Sid-80/Redis-go/store"
)

// SetKeyBase64 stores val base64-encoded under key and returns the stored value.
// Note: base64 is an *encoding*, not encryption — anyone can decode it.
func SetKeyBase64(s store.Storer, key, val string) (string, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(val))

	if err := s.Set(key, encoded); err != nil {
		return "", fmt.Errorf("set %q: %w", key, err)
	}
	return s.Get(key)
}
