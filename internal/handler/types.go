package handler

import "time"

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	SessionID				string 			`json:"session_id"`
	AccessToken 			string 			`json:"token"`
	RefreshToken 			string 			`json:"refresh_token"`
	AccessTokenExpiresAt 	time.Time 		`json:"access_token_expires_at"`
	RefreshTokenExpiresAt 	time.Time 		`json:"refresh_token_expires_at"`
	User  					UserResponse 	`json:"user"`
}

type RenewTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RenewTokenResponse struct {
	SessionID				string 			`json:"session_id"`
	AccessToken 			string 			`json:"token"`
	AccessTokenExpiresAt 	time.Time 		`json:"access_token_expires_at"`
	RefreshToken 			string 			`json:"refresh_token"`
	RefreshTokenExpiresAt 	time.Time 		`json:"refresh_token_expires_at"`
}

type UserResponse struct {
	ID 			int		`json:"id"`
	Username 	string 	`json:"username"`
	Role 		string 	`json:"role"`
}

type UserRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type UserUpdateRequest struct {
	Username string `json:"username" validate:"required"`
}

type ItemRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Price       float64 `json:"price" validate:"required,gt=0"`
}

type ItemUpdateRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Price       float64 `json:"price" validate:"required,gt=0"`
}

type ItemResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}