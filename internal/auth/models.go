package auth

import "time"

type Role string

const (
	RoleCustomer Role = "CUSTOMER"
	RoleHandyman Role = "HANDYMAN"
	RoleAdmin    Role = "ADMIN"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     Role   `json:"role,omitempty"`
}

type TokenResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
