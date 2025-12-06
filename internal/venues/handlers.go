package venues

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func userIDFromContext(c *gin.Context) (string, bool) {
	if v, ok := c.Get("uid"); ok {
		if s, ok2 := v.(string); ok2 && s != "" {
			return s, true
		}
	}
	if v, ok := c.Get("user_id"); ok {
		if s, ok2 := v.(string); ok2 && s != "" {
			return s, true
		}
	}
	return "", false
}

func List(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := userIDFromContext(c); !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		rows, err := db.Query(
			context.Background(),
			`SELECT id, owner_id, name, address, latitude, longitude,
			        avg_noise, avg_wifi, avg_crowd, created_at
			   FROM venues
			   ORDER BY created_at DESC
			   LIMIT 100`,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error"})
			return
		}
		defer rows.Close()

		var out []Venue
		for rows.Next() {
			var v Venue
			if err := rows.Scan(
				&v.ID,
				&v.OwnerID,
				&v.Name,
				&v.Address,
				&v.Latitude,
				&v.Longitude,
				&v.AvgNoise,
				&v.AvgWifi,
				&v.AvgCrowd,
				&v.CreatedAt,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "scan_error"})
				return
			}
			out = append(out, v)
		}

		c.JSON(http.StatusOK, gin.H{"venues": out})
	}
}

func Create(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := userIDFromContext(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req CreateVenueReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
			return
		}
		if req.Name == "" || req.Address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
			return
		}

		var created Venue
		err := db.QueryRow(
			context.Background(),
			`INSERT INTO venues (owner_id, name, address, latitude, longitude)
			 VALUES ($1,$2,$3,$4,$5)
			 RETURNING id, owner_id, name, address, latitude, longitude,
			           avg_noise, avg_wifi, avg_crowd, created_at`,
			uid, req.Name, req.Address, req.Latitude, req.Longitude,
		).Scan(
			&created.ID,
			&created.OwnerID,
			&created.Name,
			&created.Address,
			&created.Latitude,
			&created.Longitude,
			&created.AvgNoise,
			&created.AvgWifi,
			&created.AvgCrowd,
			&created.CreatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error"})
			return
		}

		c.JSON(http.StatusCreated, created)
	}
}
