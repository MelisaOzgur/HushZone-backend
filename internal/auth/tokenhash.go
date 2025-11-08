package auth

import (
	"crypto/sha256"
	"encoding/hex"
)

func hash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
