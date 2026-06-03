package location

import "time"

type HandymanLocation struct {
	HandymanID string    `json:"handyman_id"`
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type NearbyHandyman struct {
	HandymanID string  `json:"handyman_id"`
	DistanceKm float64 `json:"distance_km"`
}

type UpdateLocationRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
