package venues

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Venue struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	AvgNoise  float64   `json:"avg_noise"`
	AvgWifi   float64   `json:"avg_wifi"`
	AvgCrowd  float64   `json:"avg_crowd"`
	CreatedAt time.Time `json:"created_at"`
}

type createRequest struct {
	Name      string  `json:"name" binding:"required"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

type RatingRequest struct {
	Noise float64 `json:"noise"`
	Wifi  float64 `json:"wifi"`
	Crowd float64 `json:"crowd"`
}

func List(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		rows, err := db.Query(ctx, `
			SELECT
				id,
				name,
				address,
				latitude,
				longitude,
				COALESCE(avg_noise, 0),
				COALESCE(avg_wifi, 0),
				COALESCE(avg_crowd, 0),
				created_at
			FROM venues
			ORDER BY created_at DESC
			LIMIT 100`)
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
				&v.Name,
				&v.Address,
				&v.Latitude,
				&v.Longitude,
				&v.AvgNoise,
				&v.AvgWifi,
				&v.AvgCrowd,
				&v.CreatedAt,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error"})
				return
			}
			out = append(out, v)
		}
		c.JSON(http.StatusOK, gin.H{"venues": out})
	}
}

func Create(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body createRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
			return
		}

		uidVal, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, _ := uidVal.(string)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		ctx := context.Background()
		var v Venue
		err := db.QueryRow(ctx, `
			INSERT INTO venues (user_id, name, address, latitude, longitude)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING
				id,
				name,
				address,
				latitude,
				longitude,
				COALESCE(avg_noise, 0),
				COALESCE(avg_wifi, 0),
				COALESCE(avg_crowd, 0),
				created_at
		`,
			userID,
			body.Name,
			body.Address,
			body.Latitude,
			body.Longitude,
		).Scan(
			&v.ID,
			&v.Name,
			&v.Address,
			&v.Latitude,
			&v.Longitude,
			&v.AvgNoise,
			&v.AvgWifi,
			&v.AvgCrowd,
			&v.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error"})
			return
		}

		c.JSON(http.StatusCreated, v)
	}
}

func AddRating(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		venueID := c.Param("id")

		var body RatingRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
			return
		}

		uidVal, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, _ := uidVal.(string)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		ctx := context.Background()

		_, err := db.Exec(ctx, `
			INSERT INTO venue_ratings (venue_id, user_id, noise, wifi, crowd)
			VALUES ($1, $2, $3, $4, $5)
		`,
			venueID,
			userID,
			body.Noise,
			body.Wifi,
			body.Crowd,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error"})
			return
		}

		_, err = db.Exec(ctx, `
			UPDATE venues SET
				avg_noise = (SELECT COALESCE(AVG(noise), 0) FROM venue_ratings WHERE venue_id = $1),
				avg_wifi  = (SELECT COALESCE(AVG(wifi),  0) FROM venue_ratings WHERE venue_id = $1),
				avg_crowd = (SELECT COALESCE(AVG(crowd), 0) FROM venue_ratings WHERE venue_id = $1)
			WHERE id = $1
		`, venueID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"status": "rating_added"})
	}
}