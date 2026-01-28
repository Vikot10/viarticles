package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Vikot10/viarticles/internal/dto"
	"github.com/Vikot10/viarticles/internal/service/articleservice"
)

type mockArticleService struct {
	articles map[int]*dto.Article
	nextID   int
}

func newMockArticleService() *mockArticleService {
	return &mockArticleService{
		articles: make(map[int]*dto.Article),
		nextID:   1,
	}
}

func (m *mockArticleService) CreateArticle(_ context.Context, req dto.CreateArticleRequest) (*dto.Article, error) {
	if req.Title == "" {
		return nil, articleservice.ErrEmptyTitle
	}
	if req.URL == "" {
		return nil, articleservice.ErrEmptyURL
	}

	article := &dto.Article{
		ID:     m.nextID,
		Title:  req.Title,
		Body:   req.Body,
		URL:    req.URL,
		Source: req.Source,
	}
	m.articles[m.nextID] = article
	m.nextID++
	return article, nil
}

func (m *mockArticleService) GetArticleByID(_ context.Context, idStr string) (*dto.Article, error) {
	if idStr == "invalid" {
		return nil, articleservice.ErrInvalidID
	}
	id := 1
	article, ok := m.articles[id]
	if !ok {
		return nil, articleservice.ErrNotFound
	}
	return article, nil
}

func (m *mockArticleService) ListArticles(_ context.Context, _ dto.ArticleFilter) ([]dto.Article, int, error) {
	var articles []dto.Article
	for _, a := range m.articles {
		articles = append(articles, *a)
	}
	return articles, len(articles), nil
}

func (m *mockArticleService) UpdateArticle(_ context.Context, idStr string, req dto.UpdateArticleRequest) (*dto.Article, error) {
	if idStr == "invalid" {
		return nil, articleservice.ErrInvalidID
	}
	article := m.articles[1]
	if article == nil {
		return nil, articleservice.ErrNotFound
	}
	if req.Title != nil {
		article.Title = *req.Title
	}
	return article, nil
}

func (m *mockArticleService) DeleteArticle(_ context.Context, idStr string) error {
	if idStr == "invalid" {
		return articleservice.ErrInvalidID
	}
	if len(m.articles) == 0 {
		return articleservice.ErrNotFound
	}
	delete(m.articles, 1)
	return nil
}

func TestArticleHandler_Create_Success(t *testing.T) {
	svc := newMockArticleService()
	handler := NewArticleHandler(svc)

	body := dto.CreateArticleRequest{
		Title:  "Test Article",
		URL:    "https://example.com",
		Source: dto.SourceManual,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/articles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var article dto.Article
	if err := json.NewDecoder(rec.Body).Decode(&article); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if article.Title != "Test Article" {
		t.Errorf("expected title 'Test Article', got '%s'", article.Title)
	}
}

func TestArticleHandler_Create_EmptyTitle(t *testing.T) {
	svc := newMockArticleService()
	handler := NewArticleHandler(svc)

	body := dto.CreateArticleRequest{
		Title:  "",
		URL:    "https://example.com",
		Source: dto.SourceManual,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/articles", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestArticleHandler_Get_NotFound(t *testing.T) {
	svc := newMockArticleService()
	handler := NewArticleHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/articles/1", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestArticleHandler_List(t *testing.T) {
	svc := newMockArticleService()
	svc.articles[1] = &dto.Article{ID: 1, Title: "Test", URL: "https://example.com"}

	handler := NewArticleHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/articles", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestArticleHandler_Delete_Success(t *testing.T) {
	svc := newMockArticleService()
	svc.articles[1] = &dto.Article{ID: 1, Title: "Test", URL: "https://example.com"}

	handler := NewArticleHandler(svc)

	req := httptest.NewRequest(http.MethodDelete, "/articles/1", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}
