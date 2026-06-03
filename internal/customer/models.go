package customer

import "time"

type Customer struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpsertCustomerRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}
