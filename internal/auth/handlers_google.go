package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type googleReq struct {
	IDToken string `json:"id_token"`
}

type googleTokenInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"` 
	Aud           string `json:"aud"`
}

func GoogleSignIn(
	db *pgxpool.Pool,
	accessSecret, refreshSecret string,
	accessTTL, refreshTTL time.Duration,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req googleReq
		if err := c.BindJSON(&req); err != nil || strings.TrimSpace(req.IDToken) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
			return
		}

		resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + req.IDToken)
		if err != nil || resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
			return
		}
		defer resp.Body.Close()

		var info googleTokenInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil || info.Email == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
			return
		}

		wantAud := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
		if wantAud == "" || info.Aud != wantAud {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
			return
		}
		if strings.ToLower(info.EmailVerified) != "true" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "email_not_verified"})
			return
		}

		email := strings.ToLower(info.Email)
		ctx := context.Background()
		var uid string

		dummyHash, _ := HashPassword("google-" + info.Sub)
		err = db.QueryRow(ctx, `
			INSERT INTO users (email, username, password_hash, email_verified)
			VALUES ($1, split_part($1,'@',1), $2, TRUE)
			ON CONFLICT (email) DO UPDATE
				SET email_verified = TRUE,
				    updated_at = NOW()
			RETURNING id
		`, email, dummyHash).Scan(&uid)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email_exists"})
			return
		}

		access, _ := MakeToken(accessSecret, uid, accessTTL)
		refresh, _ := MakeToken(refreshSecret, uid, refreshTTL)

		_, _ = db.Exec(ctx, `
			INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
			VALUES ($1, $2, $3)
		`, uid, hash(refresh), time.Now().Add(refreshTTL))

		c.JSON(http.StatusOK, gin.H{
			"user_id": uid,
			"tokens":  TokenPair{access, refresh},
		})		
	}
}
