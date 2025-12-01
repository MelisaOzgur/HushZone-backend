package venues

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type createReq struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func List(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		rows, err := db.Query(ctx, `
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
		var req createReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
			return
		}
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name_required"})
			return
		}

		uidVal, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, _ := uidVal.(string)

		ctx := context.Background()
		var id string
		err := db.QueryRow(ctx, `
			INSERT INTO venues (name, address, latitude, longitude, created_by)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, req.Name, req.Address, req.Latitude, req.Longitude, userID).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id": id,
		})
	}
}
