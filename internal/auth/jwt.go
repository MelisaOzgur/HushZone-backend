package auth

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

func MakeToken(secret, uid string, ttl time.Duration) (string, error) {
	now := time.Now()
	cl := &Claims{
		UserID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, cl)
	return t.SignedString([]byte(secret))
}
