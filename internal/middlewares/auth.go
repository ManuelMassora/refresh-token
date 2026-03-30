package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"refresh-token/internal/auth"
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

            ctx := context.WithValue(r.Context(), "user_id", claim.ID)
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