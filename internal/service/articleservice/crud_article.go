package articleservice

import (
	"context"
	"errors"
	"strconv"

	"github.com/Vikot10/viarticles/internal/dto"
	"github.com/Vikot10/viarticles/internal/storage"
)

var (
	ErrInvalidID    = errors.New("invalid article id")
	ErrNotFound     = errors.New("article not found")
	ErrEmptyTitle   = errors.New("title is required")
	ErrEmptyURL     = errors.New("url is required")
	ErrInvalidSource = errors.New("invalid article source")
)

func (s *ArticleService) CreateArticle(ctx context.Context, req dto.CreateArticleRequest) (*dto.Article, error) {
	if req.Title == "" {
		return nil, ErrEmptyTitle
	}
	if req.URL == "" {
		return nil, ErrEmptyURL
	}
	if req.Source == "" {
		req.Source = dto.SourceManual
	}
	if !isValidSource(req.Source) {
		return nil, ErrInvalidSource
	}

	article, err := s.store.CreateArticle(ctx, req)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create article")
		return nil, err
	}

	s.logger.Info().Int("id", article.ID).Str("title", article.Title).Msg("article created")
	return article, nil
}

func (s *ArticleService) GetArticleByID(ctx context.Context, idStr string) (*dto.Article, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, ErrInvalidID
	}

	article, err := s.store.GetArticleByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.logger.Error().Err(err).Int("id", id).Msg("failed to get article")
		return nil, err
	}

	return article, nil
}

func (s *ArticleService) ListArticles(ctx context.Context, filter dto.ArticleFilter) ([]dto.Article, int, error) {
	articles, total, err := s.store.ListArticles(ctx, filter)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to list articles")
		return nil, 0, err
	}

	return articles, total, nil
}

func (s *ArticleService) UpdateArticle(ctx context.Context, idStr string, req dto.UpdateArticleRequest) (*dto.Article, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, ErrInvalidID
	}

	article, err := s.store.UpdateArticle(ctx, id, req)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.logger.Error().Err(err).Int("id", id).Msg("failed to update article")
		return nil, err
	}

	s.logger.Info().Int("id", article.ID).Msg("article updated")
	return article, nil
}

func (s *ArticleService) DeleteArticle(ctx context.Context, idStr string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ErrInvalidID
	}

	if err := s.store.DeleteArticle(ctx, id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrNotFound
		}
		s.logger.Error().Err(err).Int("id", id).Msg("failed to delete article")
		return err
	}

	s.logger.Info().Int("id", id).Msg("article deleted")
	return nil
}

func isValidSource(source dto.ArticleSource) bool {
	switch source {
	case dto.SourceManual, dto.SourceVK, dto.SourceTelegram, dto.SourceHabr:
		return true
	default:
		return false
	}
}
