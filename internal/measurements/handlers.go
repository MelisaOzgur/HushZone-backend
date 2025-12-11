package measurements

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Measurement struct {
	ID         uuid.UUID  `json:"id"`
	VenueID    uuid.UUID  `json:"venue_id"`
	UserID     uuid.UUID  `json:"user_id"`
	NoiseDB    *float64   `json:"noise_db,omitempty"`
	WifiMbps   *float64   `json:"wifi_mbps,omitempty"`
	CrowdLevel *int       `json:"crowd_level,omitempty"`
	Note       *string    `json:"note,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type createMeasurementRequest struct {
	VenueID    string   `json:"venue_id"`
	NoiseDB    *float64 `json:"noise_db"`
	WifiMbps   *float64 `json:"wifi_mbps"`
	CrowdLevel *int     `json:"crowd_level"`
	Note       *string  `json:"note"`
}

func Create(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createMeasurementRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
			return
		}

		if req.VenueID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing_venue_id"})
			return
		}

		uidVal, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userIDStr, ok := uidVal.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		venueID, err := uuid.Parse(req.VenueID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_venue_id"})
			return
		}

		var m Measurement
		row := db.QueryRow(
			c,
			`INSERT INTO measurements (user_id, venue_id, noise_db, wifi_mbps, crowd_level, note)
             VALUES ($1, $2, $3, $4, $5, $6)
             RETURNING id, venue_id, user_id, noise_db, wifi_mbps, crowd_level, note, created_at`,
			userID,
			venueID,
			req.NoiseDB,
			req.WifiMbps,
			req.CrowdLevel,
			req.Note,
		)

		err = row.Scan(
			&m.ID,
			&m.VenueID,
			&m.UserID,
			&m.NoiseDB,
			&m.WifiMbps,
			&m.CrowdLevel,
			&m.Note,
			&m.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error"})
			return
		}

		c.JSON(http.StatusCreated, m)
	}
}