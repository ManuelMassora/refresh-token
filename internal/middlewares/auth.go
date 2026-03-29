package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"refresh-token/internal/token"
	"strings"
)

func Auth(jwtMarker *token.JWTMarker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const BearerSchema = "Bearer "
			header := r.Header.Get("Authorization")
			
			if header == "" {
				renderError(w, "authorization token required", http.StatusUnauthorized)
				return
			}
			if !strings.HasPrefix(header, BearerSchema) || len(header) <= len(BearerSchema) {
                renderError(w, "Invalid authorization header format", http.StatusUnauthorized)
                return
            }
			tokenString := header[len(BearerSchema):]
            claim, err := jwtMarker.VerifyToken(tokenString)

            if err != nil {
                renderError(w, "Invalid or expired token", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), "id", claim.ID)
			ctx = context.WithValue(ctx, "username", claim.Username)
            ctx = context.WithValue(ctx, "is_admin", claim.IsAdmin)
            next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
	
func renderError(w http.ResponseWriter, message string, code int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(map[string]string{"error": message})
}