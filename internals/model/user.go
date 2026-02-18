package model

import "time"

type User struct {
	ID               string     `json:"id"`
	Email            string     `json:"email"`
	Phone            string     `json:"phone"`
	PasswordHash     string     `json:"-"`
	RefreshTokenHash string     `json:"-"`
	ResetTokenHash   string     `json:"-"`
	ResetTokenExpiry *time.Time `json:"-"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
