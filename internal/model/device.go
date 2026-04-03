package model

import "time"

type Device struct {
	ID string `gorm:"primaryKey"`
	UserID string `gorm:"index"`
	Fingerprint string `gorm:"index"`
	UserAgent string `json:"user_agent"`
	IPAddress string `json:"ip_address"`
	LastSeen time.Time `json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
}