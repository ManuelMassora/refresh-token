package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"refresh-token/internal/model"
	"refresh-token/internal/repo"
	"refresh-token/internal/util"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	repo      *repo.UserRepo
	ctx       context.Context
	validator *validator.Validate
}

func NewUserHandler(repo *repo.UserRepo, v *validator.Validate) *UserHandler {
	return &UserHandler{
		ctx:       context.Background(),
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
	user.Username = req.Username
	user.IsAdmin = req.IsAdmin
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	user.Password = hashedPassword

	createdUser, err := h.repo.CreateUser(h.ctx, &user)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(UserResponse{
		ID: createdUser.ID,
		Username: createdUser.Username,
		IsAdmin: createdUser.IsAdmin,
	})
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	uid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	user, err := h.repo.GetUserByID(h.ctx, int(uid))
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(UserResponse{
		ID: user.ID,
		Username: user.Username,
		IsAdmin: user.IsAdmin,
	})

}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	uid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.repo.GetUserByID(h.ctx, int(uid))
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	user.Username = req.Username
	err = h.repo.UpdateUser(h.ctx, user)
	if err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(UserResponse{
		ID: user.ID,
		Username: user.Username,
		IsAdmin: user.IsAdmin,
	})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	uid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	_, err = h.repo.GetUserByID(h.ctx, int(uid))
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	err = h.repo.DeleteUser(h.ctx, int(uid))
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.GetAllUsers(h.ctx)
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, UserResponse{
			ID: user.ID,
			Username: user.Username,
			IsAdmin: user.IsAdmin,
		})
	}
	json.NewEncoder(w).Encode(userResponses)
}