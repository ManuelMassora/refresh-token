package model

import "time"

type Session struct {
	UserID       string    `json:"user_id"`
	Username     string `json:"username"`
	RefreshToken string `json:"refresh_token"`
	IsRevoked    bool   `json:"is_revoked"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
}