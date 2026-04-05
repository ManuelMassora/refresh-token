package handler

import (
	"encoding/json"
	"net/http"
	"refresh-token/internal/model"
	"refresh-token/internal/repo"
	"refresh-token/internal/util"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type UserHandler struct {
	repo      *repo.UserRepo
	validator *validator.Validate
}

func NewUserHandler(repo *repo.UserRepo, v *validator.Validate) *UserHandler {
	return &UserHandler{
		repo:      repo,
		validator: v,
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	var user model.User
	user.ID = uuid.New().String()
	user.Username = req.Username
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	user.Password = hashedPassword
	user.RoleID = 2

	createdUser, err := h.repo.CreateUser(r.Context(), &user)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(UserResponse{
		ID: createdUser.ID,
		Username: createdUser.Username,
		Role: createdUser.Role.Name,
	})
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.repo.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(UserResponse{
		ID: user.ID,
		Username: user.Username,
		Role: user.Role.Name,
	})

}

func (h *UserHandler) UpdateUserName(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UserUpdateNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	err := h.repo.UpdateUserName(r.Context(), id, req.Username)
	if err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}
	user, err := h.repo.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, "User not found after update", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(UserResponse{
		ID: user.ID,
		Username: user.Username,
		Role: user.Role.Name,
	})
}

func (h *UserHandler) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UserUpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	err = h.repo.UpdateUserPassword(r.Context(), id, hashedPassword)
	if err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}
	user, err := h.repo.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, "User not found after update", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(UserResponse{
		ID: user.ID,
		Username: user.Username,
		Role: user.Role.Name,
	})
}

func (h *UserHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UserUpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	err := h.repo.UpdateUserRole(r.Context(), id, req.RoleID)
	if err != nil {
		http.Error(w, "Error updating user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := h.repo.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, "User not found after update", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(UserResponse{
		ID: user.ID,
		Username: user.Username,
		Role: user.Role.Name,
	})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := h.repo.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	err = h.repo.DeleteUser(r.Context(), id)
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.GetAllUsers(r.Context())
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, UserResponse{
			ID: user.ID,
			Username: user.Username,
			Role: user.Role.Name,
		})
	}
	json.NewEncoder(w).Encode(userResponses)
}