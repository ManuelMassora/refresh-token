package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"refresh-token/internal/auth"
	"refresh-token/internal/infra/redis"
)

func Auth(jwtMarker *auth.JWTMarker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				renderError(w, "authorization token required (cookie)", http.StatusUnauthorized)
				return
			}
			tokenString := cookie.Value

			claim, err := jwtMarker.VerifyToken(tokenString)

			if err != nil {
				renderError(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			isBlacklisted, err := redis.RedisClient.Exists(r.Context(), fmt.Sprintf("blacklist:token:%s", claim.RegisteredClaims.ID)).Result()
			if err != nil {
				log.Printf("redis unavailable during blacklist check: %v", err)
				renderError(w, "service temporarily unavailable", http.StatusServiceUnavailable)
				return
			}
			if isBlacklisted > 0 {
				renderError(w, "token has been invalidated", http.StatusUnauthorized)
				return
			}

            ctx := context.WithValue(r.Context(), auth.UserIDKey, claim.ID)
            next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
	
func renderError(w http.ResponseWriter, message string, code int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(map[string]string{"error": message})
}