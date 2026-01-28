package articleservice

import (
	"context"
	"errors"

	"github.com/Vikot10/viarticles/internal/dto"
	"github.com/Vikot10/viarticles/internal/storage"
)

var (
	ErrCategoryNotFound   = errors.New("category not found")
	ErrEmptyCategoryTitle = errors.New("category title is required")
)

func (s *ArticleService) CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (*dto.Category, error) {
	if req.Title == "" {
		return nil, ErrEmptyCategoryTitle
	}

	category, err := s.store.CreateCategory(ctx, req.Title)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create category")
		return nil, err
	}

	s.logger.Info().Int("id", category.ID).Str("title", category.Title).Msg("category created")
	return category, nil
}

func (s *ArticleService) GetCategoryByID(ctx context.Context, id int) (*dto.Category, error) {
	category, err := s.store.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrCategoryNotFound
		}
		s.logger.Error().Err(err).Int("id", id).Msg("failed to get category")
		return nil, err
	}
	return category, nil
}

func (s *ArticleService) ListCategories(ctx context.Context) ([]dto.Category, error) {
	categories, err := s.store.ListCategories(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to list categories")
		return nil, err
	}
	return categories, nil
}

func (s *ArticleService) UpdateCategory(ctx context.Context, id int, req dto.UpdateCategoryRequest) (*dto.Category, error) {
	if req.Title == "" {
		return nil, ErrEmptyCategoryTitle
	}

	category, err := s.store.UpdateCategory(ctx, id, req.Title)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrCategoryNotFound
		}
		s.logger.Error().Err(err).Int("id", id).Msg("failed to update category")
		return nil, err
	}

	s.logger.Info().Int("id", category.ID).Msg("category updated")
	return category, nil
}

func (s *ArticleService) DeleteCategory(ctx context.Context, id int) error {
	if err := s.store.DeleteCategory(ctx, id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrCategoryNotFound
		}
		s.logger.Error().Err(err).Int("id", id).Msg("failed to delete category")
		return err
	}

	s.logger.Info().Int("id", id).Msg("category deleted")
	return nil
}
