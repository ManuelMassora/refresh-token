package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"refresh-token/internal/auth"
	"refresh-token/internal/infra/redis"
	"refresh-token/internal/model"
	"refresh-token/internal/repo"
	"refresh-token/internal/util"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthHandler struct {
	repo        *repo.UserRepo
	repoSession *repo.SessionRepo
	tokenMarker *auth.JWTMarker
	validator   *validator.Validate
	DeviceRepo  *repo.DeviceRepo
}
	
func NewAuthHandler(repo *repo.UserRepo,
	repoSession *repo.SessionRepo,
	tokenMarker *auth.JWTMarker,
	v *validator.Validate,
	deviceRepo *repo.DeviceRepo) *AuthHandler {
	return &AuthHandler{
		repo:        repo,
		repoSession: repoSession,
		tokenMarker: tokenMarker,
		validator:   v,
		DeviceRepo:  deviceRepo,
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
	if err := util.CheckPasswordHash(req.Password, user.Password); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	fingerprint := generateFingerprint(user.ID, r.UserAgent())
    device, err := h.DeviceRepo.GetDeviceByFingerprint(r.Context(), fingerprint)
    if errors.Is(err, gorm.ErrRecordNotFound) {
        device = &model.Device{
            ID:          uuid.New().String(),
            UserID:      user.ID,
            Fingerprint: fingerprint,
            UserAgent:   r.UserAgent(),
            IPAddress:   getRealIP(r),
            LastSeen:    time.Now(),
        }
        if _, err := h.DeviceRepo.CreateDevice(r.Context(), device); err != nil {
            http.Error(w, "Error registering device", http.StatusInternalServerError)
            return
        }
    } else if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    } else {
        if err := h.DeviceRepo.UpdateLastSeen(r.Context(), device.ID); err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
    }

	// Invalidate previous session for this device
	oldSessionID, err := redis.RedisClient.Get(r.Context(), "device:session:"+device.ID).Result()
	if err == nil && oldSessionID != "" {
		h.repoSession.DeleteSession(r.Context(), oldSessionID)
	}

	acessToken, accessClaims, err := h.tokenMarker.CreateToken(user.ID, user.Username, user.Role.Name, 5 * time.Minute)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	refreshToken, refreshClaims, err := h.tokenMarker.CreateToken(user.ID, user.Username, user.Role.Name,  7*24*time.Hour)
	if err != nil {
		http.Error(w, "Error creating refresh token", http.StatusInternalServerError)
		return
	}

	session, err := h.repoSession.CreateSession(r.Context(), &model.Session{
		SessionID:    refreshClaims.RegisteredClaims.ID,
		UserID:       user.ID,
		Username:     refreshClaims.Username,
		RefreshToken: refreshToken,
		IsRevoked:    false,
		DeviceID:     device.ID,
		ExpiresAt:    refreshClaims.RegisteredClaims.ExpiresAt.Time,
	})
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	redis.RedisClient.Set(r.Context(), "device:session:"+device.ID, session.SessionID, 7*24*time.Hour)

	err = redis.RedisClient.Set(r.Context(), fmt.Sprintf("user:role:%s", user.ID), user.Role.Name, 5*time.Minute).Err()
	if err != nil {
		http.Error(w, "Error storing user role in Redis", http.StatusInternalServerError)
		return
	}

	h.setTokenCookies(w, session.SessionID, acessToken, refreshToken, accessClaims.RegisteredClaims.ExpiresAt.Time, refreshClaims.RegisteredClaims.ExpiresAt.Time)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func generateFingerprint(userId string, userAgent string) string {
	data := fmt.Sprintf("%s:%s", userId, userAgent)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func getRealIP(r *http.Request) string {
    // Cloudflare
    if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
        return ip
    }
    // nginx / caddy padrão
    if ip := r.Header.Get("X-Real-IP"); ip != "" {
        return ip
    }
    // fallback encadeado (pega o primeiro IP da lista)
    if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
        ip, _, _ := strings.Cut(forwarded, ",")
        return strings.TrimSpace(ip)
    }
    // desenvolvimento local — remover porta
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
    return ip
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	accessTokenCookie, err := r.Cookie("access_token")
	if err == nil {
		claims, err := h.tokenMarker.VerifyToken(accessTokenCookie.Value)
		if err == nil {
			remainingTime := time.Until(claims.ExpiresAt.Time)
			if remainingTime > 0 {
				redis.RedisClient.Set(r.Context(), fmt.Sprintf("blacklist:token:%s", claims.RegisteredClaims.ID), "blacklisted", remainingTime)
			}
		}
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}
	id := cookie.Value

	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "User ID não encontrado no contexto", http.StatusUnauthorized)
		return
	}

	session, err := h.repoSession.GetSessionByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	if session.UserID != userID {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	err = h.repoSession.DeleteSession(r.Context(), id)
	if err != nil {
		http.Error(w, "Error deleting session", http.StatusInternalServerError)
		return
	}

	if session.DeviceID != "" {
		redis.RedisClient.Del(r.Context(), "device:session:"+session.DeviceID)
	}

	redis.RedisClient.Del(r.Context(), fmt.Sprintf("user:role:%s", userID))

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
		err = h.repoSession.RevokeAllSessionsForUser(r.Context(), session.UserID)
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

	accessToken, accessClaims, err := h.tokenMarker.CreateToken(refreshClaims.ID, refreshClaims.Username, refreshClaims.Role, 5*time.Minute)
	if err != nil {
		http.Error(w, "Error creating access token", http.StatusInternalServerError)
		return
	}

	newRefreshToken, newRefreshClaims, err := h.tokenMarker.CreateToken(refreshClaims.ID, refreshClaims.Username, refreshClaims.Role, 7*24*time.Hour)
	if err != nil {
		http.Error(w, "Error creating refresh token", http.StatusInternalServerError)
		return
	}

	newSession, err := h.repoSession.CreateSession(r.Context(), &model.Session{
		SessionID:    newRefreshClaims.RegisteredClaims.ID,
		UserID:       session.UserID,
		Username:     newRefreshClaims.Username,
		RefreshToken: newRefreshToken,
		IsRevoked:    false,
		ParentID:	  session.SessionID,
		DeviceID:     session.DeviceID,
		ExpiresAt:    newRefreshClaims.RegisteredClaims.ExpiresAt.Time,
	})
	if err != nil {
		http.Error(w, "Error creating new session", http.StatusInternalServerError)
		return
	}

	if session.DeviceID != "" {
		redis.RedisClient.Set(r.Context(), "device:session:"+session.DeviceID, newSession.SessionID, 7*24*time.Hour)
	}

	err = h.repoSession.DeleteSession(r.Context(), session.SessionID)
	if err != nil {
		http.Error(w, "Error revoking old session", http.StatusInternalServerError)
		return
	}

	err = redis.RedisClient.Set(r.Context(), fmt.Sprintf("user:role:%s", refreshClaims.ID), refreshClaims.Role, 5*time.Minute).Err()
	if err != nil {
		http.Error(w, "Error updating user role in Redis", http.StatusInternalServerError)
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

	err := h.repoSession.DeleteSession(r.Context(), id)
	if err != nil {
		http.Error(w, "Error revoking session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}