package auth

import (
	"context"
	"strings"
	"time"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SignUpReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username,omitempty"`
}

type SignInReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SignInResp struct {
	Tokens TokenPair `json:"tokens"`
	UserID string    `json:"user_id,omitempty"`
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

func SignUp(db *pgxpool.Pool, accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SignUpReq
		if err := c.BindJSON(&req); err != nil || req.Email == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
			return
		}
		if !validEmail(req.Email) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_email"})
			return
		}
		if !strongPassword(req.Password) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "weak_password"})
			return
		}

		pwHash, _ := HashPassword(req.Password)

		var uid string
		err := db.QueryRow(context.Background(),
			`INSERT INTO users (email, username, password_hash) VALUES ($1,$2,$3) RETURNING id`,
			strings.ToLower(req.Email), req.Username, pwHash).Scan(&uid)
		if err != nil {
			// uniq ihlali
			c.JSON(http.StatusConflict, gin.H{"error": "email_exists"})
			return
		}

		access, _ := MakeToken(accessSecret, uid, accessTTL)
		refresh, _ := MakeToken(refreshSecret, uid, refreshTTL)
		_, _ = db.Exec(context.Background(),
			`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1,$2,$3)`,
			uid, hash(refresh), time.Now().Add(refreshTTL))

		c.JSON(http.StatusCreated, gin.H{
			"user_id": uid,
			"tokens":  TokenPair{access, refresh},
		})
	}
}

func SignIn(db *pgxpool.Pool, accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SignInReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid_json"})
			return
		}
		email := strings.TrimSpace(strings.ToLower(req.Email))
		pw := strings.TrimSpace(req.Password)
		if email == "" || pw == "" {
			c.JSON(400, gin.H{"error": "invalid_payload"})
			return
		}

		const maxFailed = 5
		const lockDuration = 10 * time.Minute

		var uid, pwHash string
		var verified bool
		var failed int
		var lockUntil *time.Time

		err := db.QueryRow(context.Background(),
			`SELECT id, password_hash, email_verified, failed_signin_attempts, lock_until
			 FROM users WHERE email=$1`, email).
			Scan(&uid, &pwHash, &verified, &failed, &lockUntil)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}

		if lockUntil != nil && lockUntil.After(time.Now()) {
			mins := int(time.Until(*lockUntil).Round(time.Minute) / time.Minute)
			if mins < 1 {
				mins = 1
			}
			c.JSON(403, gin.H{
				"error":   "account_locked",
				"minutes": mins,
			})
			return
		}

		if !verified {
			c.JSON(403, gin.H{"error": "email_not_verified"})
			return
		}

		if !VerifyPassword(pwHash, pw) {
			failed++
			if failed >= maxFailed {
				_, _ = db.Exec(context.Background(),
					`UPDATE users SET failed_signin_attempts=$1, lock_until=$2, updated_at=now() WHERE id=$3`,
					failed, time.Now().Add(lockDuration), uid)
			} else {
				_, _ = db.Exec(context.Background(),
					`UPDATE users SET failed_signin_attempts=$1, updated_at=now() WHERE id=$2`,
					failed, uid)
			}
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}

		_, _ = db.Exec(context.Background(),
			`UPDATE users SET failed_signin_attempts=0, lock_until=NULL, updated_at=now() WHERE id=$1`,
			uid)

		access, _ := MakeToken(accessSecret, uid, accessTTL)
		refresh, _ := MakeToken(refreshSecret, uid, refreshTTL)

		_, _ = db.Exec(context.Background(),
			`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1,$2,$3)`,
			uid, hash(refresh), time.Now().Add(refreshTTL))

		c.JSON(200, SignInResp{
			Tokens: TokenPair{AccessToken: access, RefreshToken: refresh},
			UserID: uid,
		})
	}
}

func Refresh(db *pgxpool.Pool, accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RefreshReq
		if err := c.BindJSON(&req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
			c.JSON(400, gin.H{"error": "invalid_json"})
			return
		}
		ref := req.RefreshToken

		var uid string
		err := db.QueryRow(context.Background(),
			`SELECT user_id FROM refresh_tokens WHERE token_hash=$1 AND expires_at > now()`,
			hash(ref)).Scan(&uid)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid_or_expired"})
			return
		}

		_, _ = db.Exec(context.Background(), `DELETE FROM refresh_tokens WHERE token_hash=$1`, hash(ref))

		access, _ := MakeToken(accessSecret, uid, accessTTL)
		newRefresh, _ := MakeToken(refreshSecret, uid, refreshTTL)
		_, _ = db.Exec(context.Background(),
			`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1,$2,$3)`,
			uid, hash(newRefresh), time.Now().Add(refreshTTL))

		c.JSON(200, gin.H{"tokens": TokenPair{AccessToken: access, RefreshToken: newRefresh}})
	}
}

func Logout(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RefreshReq
		if err := c.BindJSON(&req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
			c.JSON(400, gin.H{"error": "invalid_json"})
			return
		}
		_, _ = db.Exec(context.Background(), `DELETE FROM refresh_tokens WHERE token_hash=$1`, hash(req.RefreshToken))
		c.Status(204)
	}
}
