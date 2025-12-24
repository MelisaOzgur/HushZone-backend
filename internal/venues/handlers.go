package venues

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Venue struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Address     *string   `json:"address,omitempty"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	CreatedAt   time.Time `json:"created_at"`

	AvgNoise    *float64 `json:"avg_noise,omitempty"`
	AvgWifi     *float64 `json:"avg_wifi,omitempty"`
	AvgCrowd    *float64 `json:"avg_crowd,omitempty"`
	SampleCount int64    `json:"sample_count"`

	Source       string  `json:"source"`
	ApplePlaceID *string `json:"apple_place_id,omitempty"`
}

type createVenueReq struct {
	Name      string  `json:"name" binding:"required"`
	Address   *string `json:"address"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

type ensureVenueReq struct {
	Name         string  `json:"name" binding:"required"`
	Address      *string `json:"address"`
	Latitude     float64 `json:"latitude" binding:"required"`
	Longitude    float64 `json:"longitude" binding:"required"`
	ApplePlaceID *string `json:"apple_place_id"`
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
			  v.source,
			  v.apple_place_id,
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
			LIMIT 200
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
			return
		}
		defer rows.Close()

		out := make([]Venue, 0, 64)
		for rows.Next() {
			var v Venue
			if err := rows.Scan(
				&v.ID,
				&v.Name,
				&v.Address,
				&v.Latitude,
				&v.Longitude,
				&v.CreatedAt,
				&v.Source,
				&v.ApplePlaceID,
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
			INSERT INTO venues (user_id, name, address, latitude, longitude, source)
			VALUES ($1, $2, $3, $4, $5, 'user')
			RETURNING id, name, address, latitude, longitude, created_at, source, apple_place_id
		`, userID, req.Name, req.Address, req.Latitude, req.Longitude).Scan(
			&v.ID, &v.Name, &v.Address, &v.Latitude, &v.Longitude, &v.CreatedAt, &v.Source, &v.ApplePlaceID,
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

func Ensure(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ensureVenueReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var v Venue

		if req.ApplePlaceID != nil && strings.TrimSpace(*req.ApplePlaceID) != "" {
			appleID := strings.TrimSpace(*req.ApplePlaceID)

			err := db.QueryRow(ctx, `
				INSERT INTO venues (user_id, name, address, latitude, longitude, source, apple_place_id)
				VALUES (NULL, $1, $2, $3, $4, 'apple', $5)
				ON CONFLICT (apple_place_id)
				DO UPDATE SET
					name = EXCLUDED.name,
					address = EXCLUDED.address,
					latitude = EXCLUDED.latitude,
					longitude = EXCLUDED.longitude,
					updated_at = now()
				RETURNING id, name, address, latitude, longitude, created_at, source, apple_place_id
			`, req.Name, req.Address, req.Latitude, req.Longitude, appleID).Scan(
				&v.ID, &v.Name, &v.Address, &v.Latitude, &v.Longitude, &v.CreatedAt, &v.Source, &v.ApplePlaceID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
				return
			}

			fillVenueStats(ctx, db, &v)
			c.JSON(http.StatusOK, v)
			return
		}

		err := db.QueryRow(ctx, `
			INSERT INTO venues (user_id, name, address, latitude, longitude, source)
			VALUES (NULL, $1, $2, $3, $4, 'apple')
			ON CONFLICT DO NOTHING
			RETURNING id, name, address, latitude, longitude, created_at, source, apple_place_id
		`, req.Name, req.Address, req.Latitude, req.Longitude).Scan(
			&v.ID, &v.Name, &v.Address, &v.Latitude, &v.Longitude, &v.CreatedAt, &v.Source, &v.ApplePlaceID,
		)

		if err != nil {
			err2 := db.QueryRow(ctx, `
				SELECT id, name, address, latitude, longitude, created_at, source, apple_place_id
				FROM venues
				WHERE source='apple'
				  AND apple_place_id IS NULL
				  AND lower(name)=lower($1)
				  AND round(latitude::numeric, 5)=round($3::numeric, 5)
				  AND round(longitude::numeric, 5)=round($4::numeric, 5)
				ORDER BY created_at DESC
				LIMIT 1
			`, req.Name, req.Address, req.Latitude, req.Longitude).Scan(
				&v.ID, &v.Name, &v.Address, &v.Latitude, &v.Longitude, &v.CreatedAt, &v.Source, &v.ApplePlaceID,
			)
			if err2 != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err2.Error()})
				return
			}
		}

		fillVenueStats(ctx, db, &v)
		c.JSON(http.StatusOK, v)
	}
}

func fillVenueStats(ctx context.Context, db *pgxpool.Pool, v *Venue) {
	var avgNoise, avgWifi, avgCrowd *float64
	var n int64

	_ = db.QueryRow(ctx, `
		SELECT
		  AVG(m.noise_db) AS avg_noise,
		  AVG(m.wifi_mbps) AS avg_wifi,
		  AVG(m.crowd_level) AS avg_crowd,
		  COUNT(m.id) AS sample_count
		FROM measurements m
		WHERE m.venue_id = $1
		  AND m.created_at >= now() - interval '30 minutes'
	`, v.ID).Scan(&avgNoise, &avgWifi, &avgCrowd, &n)

	v.AvgNoise = avgNoise
	v.AvgWifi = avgWifi
	v.AvgCrowd = avgCrowd
	v.SampleCount = n
}