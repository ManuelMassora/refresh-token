package middlewares

import (
	"net/http"
	"refresh-token/internal/auth"
	"refresh-token/internal/handler"
)

func HasAnyRole(h *handler.AuthHandler, allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(auth.UserIDKey).(string)
			if !ok {
				renderError(w, "Erro interno de identificação", http.StatusInternalServerError)
				return
			}
			actualRole, err := h.GetOrFetchUserRole(r.Context(), userID)
			if err != nil {
				renderError(w, "Permissão não encontrada", http.StatusForbidden)
				return
			}
			for _, role := range allowedRoles {
				if actualRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			renderError(w, "Acesso negado para o seu nível de usuário", http.StatusForbidden)
		})
	}
}
