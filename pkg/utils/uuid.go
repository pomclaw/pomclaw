package utils

import (
	"crypto/rand"
	"strings"

	"github.com/oklog/ulid/v2"
)

// GenerateID generates a new ULID (26 characters, time-ordered)
func GenerateID() string {
	id, _ := ulid.New(ulid.Now(), rand.Reader)
	return id.String()
}

func GenerateShortID() string {
	return strings.ToLower(GenerateID()[:8])
}
