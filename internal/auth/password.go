package auth

import "github.com/alexedwards/argon2id"

func HashPassword(pw string) (string, error) {
	return argon2id.CreateHash(pw, argon2id.DefaultParams)
}

func VerifyPassword(hash, pw string) bool {
	ok, _ := argon2id.ComparePasswordAndHash(pw, hash)
	return ok
}
