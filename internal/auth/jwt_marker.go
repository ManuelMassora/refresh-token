package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTMarker struct {
	secretKey string
}

func NewJWTMarker(secretKey string) *JWTMarker {
	return &JWTMarker{secretKey: secretKey}
}

func (marker *JWTMarker) CreateToken(id string, username string, role string, duration time.Duration) (string, *UserClaims, error) {
	claims, err := NewUserClaims(id, username, role, duration)
	if err != nil {
		return "", nil, err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(marker.secretKey))
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, claims, nil
}

func (marker *JWTMarker) VerifyToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(marker.secretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("error extracting claims: %w", err)
	}
	return claims, nil
}