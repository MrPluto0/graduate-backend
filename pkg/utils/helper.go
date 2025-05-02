package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomString generates a random string of the specified length.
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsEmpty checks if a string is empty.
func IsEmpty(s string) bool {
	return len(s) == 0
}
