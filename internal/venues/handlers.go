package venues

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Venue struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Address     *string    `json:"address,omitempty"`
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
	CreatedAt   time.Time  `json:"created_at"`
	AvgNoise    *float64   `json:"avg_noise,omitempty"`
	AvgWifi     *float64   `json:"avg_wifi,omitempty"`
	AvgCrowd    *float64   `json:"avg_crowd,omitempty"`
	SampleCount int64      `json:"sample_count"`
}

type createVenueReq struct {
	Name      string  `json:"name" binding:"required"`
	Address   *string `json:"address"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

func List(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		rows, err := db.Query(ctx, `
			SELECT
			  v.id,
			  v.name,
			  v.address,
			  v.latitude,
			  v.longitude,
			  v.created_at,
			  AVG(m.noise_db) AS avg_noise,
			  AVG(m.wifi_mbps) AS avg_wifi,
			  AVG(m.crowd_level) AS avg_crowd,
			  COUNT(m.id) AS sample_count
			FROM venues v
			LEFT JOIN measurements m
			  ON m.venue_id = v.id
			 AND m.created_at >= now() - interval '30 minutes'
			GROUP BY v.id
			ORDER BY v.created_at DESC
			LIMIT 100
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
			return
		}
		defer rows.Close()

		out := make([]Venue, 0, 32)
		for rows.Next() {
			var v Venue
			if err := rows.Scan(
				&v.ID,
				&v.Name,
				&v.Address,
				&v.Latitude,
				&v.Longitude,
				&v.CreatedAt,
				&v.AvgNoise,
				&v.AvgWifi,
				&v.AvgCrowd,
				&v.SampleCount,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
				return
			}
			out = append(out, v)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"venues": out})
	}
}

func Create(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidVal, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, ok := uidVal.(string)
		if !ok || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req createVenueReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var v Venue
		err := db.QueryRow(ctx, `
			INSERT INTO venues (user_id, name, address, latitude, longitude)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, name, address, latitude, longitude, created_at
		`, userID, req.Name, req.Address, req.Latitude, req.Longitude).Scan(
			&v.ID, &v.Name, &v.Address, &v.Latitude, &v.Longitude, &v.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
			return
		}

		v.AvgNoise = nil
		v.AvgWifi = nil
		v.AvgCrowd = nil
		v.SampleCount = 0

		c.JSON(http.StatusCreated, v)
	}
}