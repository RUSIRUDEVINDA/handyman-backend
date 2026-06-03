package handyman

import "time"

type AvailabilityStatus string

const (
	StatusAvailable AvailabilityStatus = "AVAILABLE"
	StatusBusy      AvailabilityStatus = "BUSY"
	StatusOffline   AvailabilityStatus = "OFFLINE"
)

type Handyman struct {
	ID          string             `json:"id"`
	UserID      string             `json:"user_id"`
	FullName    string             `json:"full_name"`
	Phone       string             `json:"phone"`
	Bio         string             `json:"bio"`
	Skills      []string           `json:"skills"`
	HourlyRate  int                `json:"hourly_rate"`
	Status      AvailabilityStatus `json:"status"`
	Rating      float64            `json:"rating"`
	ReviewCount int                `json:"review_count"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type UpsertHandymanRequest struct {
	FullName   string             `json:"full_name"`
	Phone      string             `json:"phone"`
	Bio        string             `json:"bio"`
	Skills     []string           `json:"skills"`
	HourlyRate int                `json:"hourly_rate"`
	Status     AvailabilityStatus `json:"status"`
}

type SearchRequest struct {
	Skill string `query:"skill"`
}
