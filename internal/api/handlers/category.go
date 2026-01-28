package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Vikot10/viarticles/internal/api"
	"github.com/Vikot10/viarticles/internal/dto"
	"github.com/Vikot10/viarticles/internal/service/articleservice"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (*dto.Category, error)
	GetCategoryByID(ctx context.Context, id int) (*dto.Category, error)
	ListCategories(ctx context.Context) ([]dto.Category, error)
	UpdateCategory(ctx context.Context, id int, req dto.UpdateCategoryRequest) (*dto.Category, error)
	DeleteCategory(ctx context.Context, id int) error
}

type CategoryHandler struct {
	svc CategoryService
}

func NewCategoryHandler(svc CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid json body")
		return
	}

	category, err := h.svc.CreateCategory(r.Context(), req)
	if err != nil {
		if errors.Is(err, articleservice.ErrEmptyCategoryTitle) {
			api.BadRequest(w, "title is required")
			return
		}
		api.InternalError(w)
		return
	}

	api.JSON(w, http.StatusCreated, category)
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		api.BadRequest(w, "invalid category id")
		return
	}

	category, err := h.svc.GetCategoryByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, articleservice.ErrCategoryNotFound) {
			api.NotFound(w, "category not found")
			return
		}
		api.InternalError(w)
		return
	}

	api.JSON(w, http.StatusOK, category)
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.ListCategories(r.Context())
	if err != nil {
		api.InternalError(w)
		return
	}

	if categories == nil {
		categories = []dto.Category{}
	}

	api.JSON(w, http.StatusOK, categories)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		api.BadRequest(w, "invalid category id")
		return
	}

	var req dto.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid json body")
		return
	}

	category, err := h.svc.UpdateCategory(r.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, articleservice.ErrCategoryNotFound):
			api.NotFound(w, "category not found")
		case errors.Is(err, articleservice.ErrEmptyCategoryTitle):
			api.BadRequest(w, "title is required")
		default:
			api.InternalError(w)
		}
		return
	}

	api.JSON(w, http.StatusOK, category)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		api.BadRequest(w, "invalid category id")
		return
	}

	if err := h.svc.DeleteCategory(r.Context(), id); err != nil {
		if errors.Is(err, articleservice.ErrCategoryNotFound) {
			api.NotFound(w, "category not found")
			return
		}
		api.InternalError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
