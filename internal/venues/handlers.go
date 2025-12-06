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

func List(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(
			context.Background(),
			`SELECT id, name, address, latitude, longitude
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
				&v.Name,
				&v.Address,
				&v.Latitude,
				&v.Longitude,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error"})
				return
			}
			out = append(out, v)
		}

		c.JSON(http.StatusOK, gin.H{"venues": out})
	}
}

type createReq struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func Create(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createReq
		if err := c.BindJSON(&req); err != nil ||
			req.Name == "" || req.Address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
			return
		}

		var id string
		err := db.QueryRow(
			context.Background(),
			`INSERT INTO venues (name, address, latitude, longitude)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id`,
			req.Name, req.Address, req.Latitude, req.Longitude,
		).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":        id,
			"name":      req.Name,
			"address":   req.Address,
			"latitude":  req.Latitude,
			"longitude": req.Longitude,
		})
	}
}
