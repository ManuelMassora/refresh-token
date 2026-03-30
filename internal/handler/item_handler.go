package handler

import (
	"encoding/json"
	"net/http"
	"refresh-token/internal/model"
	"refresh-token/internal/repo"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type ItemHandler struct {
	repo      *repo.ItemRepo
	validator *validator.Validate
}

func NewItemHandler(repo *repo.ItemRepo, v *validator.Validate) *ItemHandler {
	return &ItemHandler{
		repo:      repo,
		validator: v,
	}
}

func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req ItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	item := model.Item{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	}

	createdItem, err := h.repo.CreateItem(r.Context(), &item)
	if err != nil {
		http.Error(w, "Error creating item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdItem)
}

func (h *ItemHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	uid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := h.repo.GetItemByID(r.Context(), int(uid))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(item)
}

func (h *ItemHandler) GetAllItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.GetAllItems(r.Context())
	if err != nil {
		http.Error(w, "Error fetching items", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(items)
}

func (h *ItemHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	uid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var req ItemUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	item, err := h.repo.GetItemByID(r.Context(), int(uid))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	item.Name = req.Name
	item.Description = req.Description
	item.Price = req.Price

	if err := h.repo.UpdateItem(r.Context(), item); err != nil {
		http.Error(w, "Error updating item", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(item)
}

func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	uid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	_, err = h.repo.GetItemByID(r.Context(), int(uid))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	if err := h.repo.DeleteItem(r.Context(), int(uid)); err != nil {
		http.Error(w, "Error deleting item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
