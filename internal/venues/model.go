package venues

import "time"

type Venue struct {
	ID        string     `json:"id"`
	OwnerID   string     `json:"owner_id"`
	Name      string     `json:"name"`
	Address   string     `json:"address"`
	Latitude  float64    `json:"latitude"`
	Longitude float64    `json:"longitude"`
	AvgNoise  *float64   `json:"avg_noise,omitempty"`
	AvgWifi   *float64   `json:"avg_wifi,omitempty"`
	AvgCrowd  *float64   `json:"avg_crowd,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateVenueReq struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
