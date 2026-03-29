package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"refresh-token/internal/model"
	"refresh-token/internal/repo"
	"refresh-token/internal/token"
	"refresh-token/internal/util"
	"time"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	repo *repo.UserRepo
	repoSession *repo.SessionRepo
	tokenMarker *token.JWTMarker
	ctx context.Context
}
	
func NewAuthHandler(repo *repo.UserRepo, repoSession *repo.SessionRepo, tokenMarker *token.JWTMarker) *AuthHandler {
	return &AuthHandler{
		repo: repo,
		repoSession: repoSession,
		tokenMarker: tokenMarker,
		ctx: context.Background(),
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByUsername(h.ctx, req.Username)
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

	session, err := h.repoSession.CreateSession(h.ctx, &model.Session{
		UserID: refreshClaims.RegisteredClaims.ID,
		Username: refreshClaims.Username,
		RefreshToken: refreshToken,
		IsRevoked: false,
		ExpiresAt: refreshClaims.RegisteredClaims.ExpiresAt.Time,
	})
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	res := LoginResponse{
		SessionID: session.UserID,
		AccessToken: acessToken,
		RefreshToken: refreshToken,
		AccessTokenExpiresAt: accessClaims.RegisteredClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: refreshClaims.RegisteredClaims.ExpiresAt.Time,
		User: UserResponse{
			ID: user.ID,
			Username: user.Username,
			IsAdmin: user.IsAdmin,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	err := h.repoSession.DeleteSession(h.ctx, id)
	if err != nil {
		http.Error(w, "Error deleting session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) RenewAccessToken(w http.ResponseWriter, r *http.Request) {
	var req RenewTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	refreshClaims, err := h.tokenMarker.VerifyToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token: " + err.Error(), http.StatusUnauthorized)
		return
	}

	session, err := h.repoSession.GetSessionByID(h.ctx, refreshClaims.RegisteredClaims.ID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	if session.IsRevoked {
		http.Error(w, "Session revoked", http.StatusUnauthorized)
		return
	}

	if session.Username != refreshClaims.Username {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	acessToken, acessClaims, err := h.tokenMarker.CreateToken(refreshClaims.ID, refreshClaims.Username, refreshClaims.IsAdmin, 2*time.Minute)
	if err != nil {
		http.Error(w, "Error creating access token", http.StatusInternalServerError)
		return
	}

	res := RenewTokenResponse{
		AccessToken: acessToken,
		AccessTokenExpiresAt: acessClaims.RegisteredClaims.ExpiresAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	err := h.repoSession.RevokeSession(h.ctx, id)
	if err != nil {
		http.Error(w, "Error revoking session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}