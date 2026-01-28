package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/Vikot10/viarticles/internal/dto"
)

func (s *Storage) CreateCategory(ctx context.Context, title string) (*dto.Category, error) {
	var category dto.Category
	err := s.pg.QueryRow(ctx, `
		INSERT INTO category (title)
		VALUES ($1)
		ON CONFLICT (title) DO UPDATE SET title = EXCLUDED.title
		RETURNING id, title, created_at, updated_at
	`, title).Scan(&category.ID, &category.Title, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *Storage) GetCategoryByID(ctx context.Context, id int) (*dto.Category, error) {
	var category dto.Category
	err := s.pg.QueryRow(ctx, `
		SELECT id, title, created_at, updated_at
		FROM category
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&category.ID, &category.Title, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &category, nil
}

func (s *Storage) GetCategoryByTitle(ctx context.Context, title string) (*dto.Category, error) {
	var category dto.Category
	err := s.pg.QueryRow(ctx, `
		SELECT id, title, created_at, updated_at
		FROM category
		WHERE title = $1 AND deleted_at IS NULL
	`, title).Scan(&category.ID, &category.Title, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &category, nil
}

func (s *Storage) ListCategories(ctx context.Context) ([]dto.Category, error) {
	rows, err := s.pg.Query(ctx, `
		SELECT id, title, created_at, updated_at
		FROM category
		WHERE deleted_at IS NULL
		ORDER BY title
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []dto.Category
	for rows.Next() {
		var category dto.Category
		if err := rows.Scan(&category.ID, &category.Title, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (s *Storage) UpdateCategory(ctx context.Context, id int, title string) (*dto.Category, error) {
	var category dto.Category
	err := s.pg.QueryRow(ctx, `
		UPDATE category
		SET title = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, title, created_at, updated_at
	`, id, title).Scan(&category.ID, &category.Title, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &category, nil
}

func (s *Storage) DeleteCategory(ctx context.Context, id int) error {
	result, err := s.pg.Exec(ctx, `
		UPDATE category
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Storage) GetOrCreateCategories(ctx context.Context, titles []string) ([]dto.Category, error) {
	if len(titles) == 0 {
		return nil, nil
	}

	categories := make([]dto.Category, 0, len(titles))
	for _, title := range titles {
		cat, err := s.CreateCategory(ctx, title)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *cat)
	}
	return categories, nil
}
