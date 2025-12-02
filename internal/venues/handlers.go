package venues

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Venue struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CreateVenueReq struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func userIDFromCtx(c *gin.Context) (string, bool) {
	// En çok kullanılan isimler:
	if v, ok := c.Get("user_id"); ok {
		if s, ok2 := v.(string); ok2 {
			return s, true
		}
	}
	if v, ok := c.Get("uid"); ok {
		if s, ok2 := v.(string); ok2 {
			return s, true
		}
	}
	if v, ok := c.Get("userID"); ok {
		if s, ok2 := v.(string); ok2 {
			return s, true
		}
	}
	return "", false
}

func List(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(context.Background(), `
            SELECT id, name, address, latitude, longitude
            FROM venues
            ORDER BY created_at DESC
            LIMIT 100
        `)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error"})
			return
		}
		defer rows.Close()

		var items []Venue
		for rows.Next() {
			var v Venue
			if err := rows.Scan(&v.ID, &v.Name, &v.Address, &v.Latitude, &v.Longitude); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "scan_error"})
				return
			}
			items = append(items, v)
		}

		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}

func Create(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateVenueReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
			return
		}
		if req.Name == "" || req.Address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing_fields"})
			return
		}

		userID, ok := userIDFromCtx(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var newID string
		err := db.QueryRow(
			context.Background(),
			`INSERT INTO venues (user_id, name, address, latitude, longitude)
             VALUES ($1, $2, $3, $4, $5)
             RETURNING id`,
			userID, req.Name, req.Address, req.Latitude, req.Longitude,
		).Scan(&newID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_insert_error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": newID})
	}
}
