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

type ArticleService interface {
	CreateArticle(ctx context.Context, req dto.CreateArticleRequest) (*dto.Article, error)
	GetArticleByID(ctx context.Context, idStr string) (*dto.Article, error)
	ListArticles(ctx context.Context, filter dto.ArticleFilter) ([]dto.Article, int, error)
	UpdateArticle(ctx context.Context, idStr string, req dto.UpdateArticleRequest) (*dto.Article, error)
	DeleteArticle(ctx context.Context, idStr string) error
}

type ArticleHandler struct {
	svc ArticleService
}

func NewArticleHandler(svc ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
}

func (h *ArticleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid json body")
		return
	}

	article, err := h.svc.CreateArticle(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, articleservice.ErrEmptyTitle):
			api.BadRequest(w, "title is required")
		case errors.Is(err, articleservice.ErrEmptyURL):
			api.BadRequest(w, "url is required")
		case errors.Is(err, articleservice.ErrInvalidSource):
			api.BadRequest(w, "invalid source")
		default:
			api.InternalError(w)
		}
		return
	}

	api.JSON(w, http.StatusCreated, article)
}

func (h *ArticleHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	article, err := h.svc.GetArticleByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, articleservice.ErrInvalidID):
			api.BadRequest(w, "invalid article id")
		case errors.Is(err, articleservice.ErrNotFound):
			api.NotFound(w, "article not found")
		default:
			api.InternalError(w)
		}
		return
	}

	api.JSON(w, http.StatusOK, article)
}

func (h *ArticleHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := dto.ArticleFilter{
		Limit:  20,
		Offset: 0,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	if source := r.URL.Query().Get("source"); source != "" {
		s := dto.ArticleSource(source)
		filter.Source = &s
	}

	if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
		if categoryID, err := strconv.Atoi(categoryIDStr); err == nil {
			filter.CategoryID = &categoryID
		}
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = &search
	}

	articles, total, err := h.svc.ListArticles(r.Context(), filter)
	if err != nil {
		api.InternalError(w)
		return
	}

	if articles == nil {
		articles = []dto.Article{}
	}

	api.JSON(w, http.StatusOK, api.ListResponse[dto.Article]{
		Data:   articles,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
}

func (h *ArticleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid json body")
		return
	}

	article, err := h.svc.UpdateArticle(r.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, articleservice.ErrInvalidID):
			api.BadRequest(w, "invalid article id")
		case errors.Is(err, articleservice.ErrNotFound):
			api.NotFound(w, "article not found")
		default:
			api.InternalError(w)
		}
		return
	}

	api.JSON(w, http.StatusOK, article)
}

func (h *ArticleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.DeleteArticle(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, articleservice.ErrInvalidID):
			api.BadRequest(w, "invalid article id")
		case errors.Is(err, articleservice.ErrNotFound):
			api.NotFound(w, "article not found")
		default:
			api.InternalError(w)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
