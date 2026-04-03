package model

import "time"

type Session struct {
	SessionID       string    `json:"session_id"`
	UserID       	string `json:"user_id"`
	Username     	string `json:"username"`
	RefreshToken 	string `json:"refresh_token"`
	IsRevoked    	bool   `json:"is_revoked"`
	ParentID		string	`json:"parent_id"`
	DeviceID		string	`json:"device_id"`
	CreatedAt    	time.Time  `json:"created_at"`
	ExpiresAt    	time.Time  `json:"expires_at"`
}