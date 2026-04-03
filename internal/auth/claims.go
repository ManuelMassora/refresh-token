package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ContextKey string

const (
	UserIDKey ContextKey = "user_id"
)

type UserClaims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role  		string   `json:"role"`
	jwt.RegisteredClaims
}

func NewUserClaims(id string, username string, role string, duration time.Duration) (*UserClaims, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token ID: %w", err)
	}
	return &UserClaims{
		ID:       id,
		Username: username,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: tokenID.String(),
			Subject: username,
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}, nil
}