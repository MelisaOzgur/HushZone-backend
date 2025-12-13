package measurements

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type createReq struct {
	VenueID    string   `json:"venue_id" binding:"required"`
	NoiseDB    *float64 `json:"noise_db"`
	WifiMbps   *float64 `json:"wifi_mbps"`
	CrowdLevel *int     `json:"crowd_level"`
	Note       *string  `json:"note"`
}

type Measurement struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	VenueID    string    `json:"venue_id"`
	NoiseDB    *float64  `json:"noise_db,omitempty"`
	WifiMbps   *float64  `json:"wifi_mbps,omitempty"`
	CrowdLevel *int      `json:"crowd_level,omitempty"`
	Note       *string   `json:"note,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
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

		var req createReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var m Measurement
		err := db.QueryRow(ctx, `
			INSERT INTO measurements (user_id, venue_id, noise_db, wifi_mbps, crowd_level, note)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, user_id, venue_id, noise_db, wifi_mbps, crowd_level, note, created_at
		`, userID, req.VenueID, req.NoiseDB, req.WifiMbps, req.CrowdLevel, req.Note).Scan(
			&m.ID, &m.UserID, &m.VenueID, &m.NoiseDB, &m.WifiMbps, &m.CrowdLevel, &m.Note, &m.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, m)
	}
}