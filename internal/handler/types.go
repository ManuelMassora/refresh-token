package handler

import "time"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	SessionID				string 			`json:"session_id"`
	AccessToken 			string 			`json:"token"`
	RefreshToken 			string 			`refresh_token"`
	AccessTokenExpiresAt 	time.Time 		`json:"access_token_expires_at"`
	RefreshTokenExpiresAt 	time.Time 		`json:"refresh_token_expires_at"`
	User  					UserResponse 	`json:"user"`
}

type RenewTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RenewTokenResponse struct {
	AccessToken 			string 			`json:"token"`
	AccessTokenExpiresAt 	time.Time 		`json:"access_token_expires_at"`
}

type UserResponse struct {
	ID 			int		`json:"id"`
	Username 	string 	`json:"username"`
	IsAdmin 	bool 	`json:"is_admin"`
}

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

type UserUpdateRequest struct {
	Username string `json:"username"`
}

type ItemRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type ItemUpdateRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type ItemResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}