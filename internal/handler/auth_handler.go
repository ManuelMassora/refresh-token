package handler

import (
	"encoding/json"
	"net/http"
	"refresh-token/internal/auth"
	"refresh-token/internal/model"
	"refresh-token/internal/repo"
	"refresh-token/internal/util"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	repo        *repo.UserRepo
	repoSession *repo.SessionRepo
	tokenMarker *auth.JWTMarker
	validator   *validator.Validate
}
	
func NewAuthHandler(repo *repo.UserRepo, repoSession *repo.SessionRepo, tokenMarker *auth.JWTMarker, v *validator.Validate) *AuthHandler {
	return &AuthHandler{
		repo:        repo,
		repoSession: repoSession,
		tokenMarker: tokenMarker,
		validator:   v,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	if err != util.CheckPasswordHash(req.Password, user.Password) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	acessToken, accessClaims, err := h.tokenMarker.CreateToken(int64(user.ID), user.Username, user.IsAdmin, 2 * time.Minute)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	refreshToken, refreshClaims, err := h.tokenMarker.CreateToken(int64(user.ID), user.Username, user.IsAdmin, 6 * time.Minute)
	if err != nil {
		http.Error(w, "Error creating refresh token", http.StatusInternalServerError)
		return
	}

	session, err := h.repoSession.CreateSession(r.Context(), &model.Session{
		SessionID: refreshClaims.RegisteredClaims.ID,
		UserID: user.ID,
		Username: refreshClaims.Username,
		RefreshToken: refreshToken,
		IsRevoked: false,
		ExpiresAt: refreshClaims.RegisteredClaims.ExpiresAt.Time,
	})
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	h.setTokenCookies(w, session.SessionID, acessToken, refreshToken, accessClaims.RegisteredClaims.ExpiresAt.Time, refreshClaims.RegisteredClaims.ExpiresAt.Time)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}
	id := cookie.Value

	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "User ID não encontrado no contexto", http.StatusUnauthorized)
		return
	}

	session, err := h.repoSession.GetSessionByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	if int64(session.UserID) != userID {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	err = h.repoSession.DeleteSession(r.Context(), id)
	if err != nil {
		http.Error(w, "Error deleting session", http.StatusInternalServerError)
		return
	}

	h.clearTokenCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) setTokenCookies(w http.ResponseWriter, sessionID, accessToken, refreshToken string, accessExpires, refreshExpires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  refreshExpires,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  accessExpires,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  refreshExpires,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearTokenCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) RenewAccessToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Refresh token required (cookie)", http.StatusUnauthorized)
		return
	}
	refreshToken := cookie.Value

	refreshClaims, err := h.tokenMarker.VerifyToken(refreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	session, err := h.repoSession.GetSessionByID(r.Context(), refreshClaims.RegisteredClaims.ID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	if session.IsRevoked {
		err = h.repoSession.RevokeAllSessionForUser(r.Context(), session.UserID)
		if err != nil {
			http.Error(w, "Error revoking session", http.StatusInternalServerError)
			return
		}
		http.Error(w, "Session revoked, all sessions revoked", http.StatusUnauthorized)
		return
	}

	if session.Username != refreshClaims.Username {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	if session.RefreshToken != refreshToken {
		http.Error(w, "Mismatched refresh token", http.StatusUnauthorized)
		return
	}

	accessToken, accessClaims, err := h.tokenMarker.CreateToken(refreshClaims.ID, refreshClaims.Username, refreshClaims.IsAdmin, 2*time.Minute)
	if err != nil {
		http.Error(w, "Error creating access token", http.StatusInternalServerError)
		return
	}

	newRefreshToken, newRefreshClaims, err := h.tokenMarker.CreateToken(refreshClaims.ID, refreshClaims.Username, refreshClaims.IsAdmin, 6*time.Minute)
	if err != nil {
		http.Error(w, "Error creating refresh token", http.StatusInternalServerError)
		return
	}

	_, err = h.repoSession.CreateSession(r.Context(), &model.Session{
		SessionID:    newRefreshClaims.RegisteredClaims.ID,
		UserID:       session.UserID,
		Username:     newRefreshClaims.Username,
		RefreshToken: newRefreshToken,
		IsRevoked:    false,
		ParentID:	  session.SessionID,
		ExpiresAt:    newRefreshClaims.RegisteredClaims.ExpiresAt.Time,
	})
	if err != nil {
		http.Error(w, "Error creating new session", http.StatusInternalServerError)
		return
	}

	err = h.repoSession.RevokeSession(r.Context(), session.SessionID)
	if err != nil {
		http.Error(w, "Error revoking old session", http.StatusInternalServerError)
		return
	}

	h.setTokenCookies(w, newRefreshClaims.RegisteredClaims.ID, accessToken, newRefreshToken, accessClaims.RegisteredClaims.ExpiresAt.Time, newRefreshClaims.RegisteredClaims.ExpiresAt.Time)

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	err := h.repoSession.RevokeSession(r.Context(), id)
	if err != nil {
		http.Error(w, "Error revoking session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}