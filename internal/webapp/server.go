package webapp

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Vikot10/viarticles/internal/storage"
)

//go:embed templates/*.html
var templateFiles embed.FS

type Server struct {
	store     *storage.Storage
	templates *template.Template
}

type pageData struct {
	Articles []storage.Article
	Filter   string
	Search   string
	Error    string
}

func New(store *storage.Storage) (*Server, error) {
	templates, err := template.ParseFS(templateFiles, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	return &Server{store: store, templates: templates}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /", s.index)
	mux.HandleFunc("POST /articles", s.createArticle)
	mux.HandleFunc("POST /articles/{id}/read", s.toggleRead)
	mux.HandleFunc("POST /articles/{id}/like", s.toggleLiked)
	mux.HandleFunc("POST /articles/{id}/delete", s.deleteArticle)
	return logging(mux)
}

func (s *Server) health(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("ok"))
}

func (s *Server) index(writer http.ResponseWriter, request *http.Request) {
	data, err := s.loadPage(request)
	if err != nil {
		http.Error(writer, "Не удалось загрузить статьи", http.StatusInternalServerError)
		return
	}
	s.render(writer, "page", data)
}

func (s *Server) createArticle(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		s.mutationError(writer, request, "Некорректная форма", http.StatusBadRequest)
		return
	}
	title := strings.TrimSpace(request.FormValue("title"))
	articleURL := strings.TrimSpace(request.FormValue("url"))
	body := strings.TrimSpace(request.FormValue("body"))
	if err := validateArticle(title, articleURL); err != nil {
		s.mutationError(writer, request, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := s.store.CreateArticle(request.Context(), storage.NewArticle{
		Title: title, Body: body, URL: articleURL, Source: "manual",
	}); err != nil {
		s.mutationError(writer, request, "Не удалось сохранить статью", http.StatusInternalServerError)
		return
	}
	s.respondAfterMutation(writer, request)
}

func (s *Server) toggleRead(writer http.ResponseWriter, request *http.Request) {
	id, ok := articleID(writer, request)
	if !ok {
		return
	}
	article, err := s.store.ToggleRead(request.Context(), id)
	if err != nil {
		s.handleArticleError(writer, err)
		return
	}
	s.respondWithArticle(writer, request, article)
}

func (s *Server) toggleLiked(writer http.ResponseWriter, request *http.Request) {
	id, ok := articleID(writer, request)
	if !ok {
		return
	}
	article, err := s.store.ToggleLiked(request.Context(), id)
	if err != nil {
		s.handleArticleError(writer, err)
		return
	}
	s.respondWithArticle(writer, request, article)
}

func (s *Server) deleteArticle(writer http.ResponseWriter, request *http.Request) {
	id, ok := articleID(writer, request)
	if !ok {
		return
	}
	if err := s.store.DeleteArticle(request.Context(), id); err != nil {
		s.handleArticleError(writer, err)
		return
	}
	if request.Header.Get("HX-Request") == "true" {
		writer.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(writer, request, "/", http.StatusSeeOther)
}

func (s *Server) respondWithArticle(writer http.ResponseWriter, request *http.Request, article storage.Article) {
	if request.Header.Get("HX-Request") == "true" {
		s.render(writer, "article", article)
		return
	}
	http.Redirect(writer, request, "/", http.StatusSeeOther)
}

func (s *Server) respondAfterMutation(writer http.ResponseWriter, request *http.Request) {
	if request.Header.Get("HX-Request") != "true" {
		http.Redirect(writer, request, "/", http.StatusSeeOther)
		return
	}
	data, err := s.loadPage(request)
	if err != nil {
		http.Error(writer, "Не удалось обновить список", http.StatusInternalServerError)
		return
	}
	s.render(writer, "article-list", data)
}

func (s *Server) mutationError(writer http.ResponseWriter, request *http.Request, message string, status int) {
	if request.Header.Get("HX-Request") == "true" {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(status)
		_, _ = fmt.Fprintf(writer, `<div class="error">%s</div>`, template.HTMLEscapeString(message))
		return
	}
	http.Error(writer, message, status)
}

func (s *Server) loadPage(request *http.Request) (pageData, error) {
	filter := strings.TrimSpace(request.FormValue("filter"))
	search := strings.TrimSpace(request.FormValue("search"))
	articles, err := s.store.ListArticles(request.Context(), filter, search)
	return pageData{Articles: articles, Filter: filter, Search: search}, err
}

func (s *Server) handleArticleError(writer http.ResponseWriter, err error) {
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(writer, "Статья не найдена", http.StatusNotFound)
		return
	}
	http.Error(writer, "Ошибка хранилища", http.StatusInternalServerError)
}

func (s *Server) render(writer http.ResponseWriter, name string, value any) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(writer, name, value); err != nil {
		slog.Error("render template", "name", name, "error", err)
	}
}

func articleID(writer http.ResponseWriter, request *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.Error(writer, "Некорректный ID", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func validateArticle(title, value string) error {
	if title == "" {
		return errors.New("Укажите название")
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("Укажите корректную http(s)-ссылку")
	}
	return nil
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		slog.Info("http request", "method", request.Method, "path", request.URL.Path)
		next.ServeHTTP(writer, request)
	})
}
