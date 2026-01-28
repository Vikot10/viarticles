package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/Vikot10/viarticles/internal/dto"
)

func (s *Storage) CreateArticle(ctx context.Context, req dto.CreateArticleRequest) (*dto.Article, error) {
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var article dto.Article
	err = tx.QueryRow(ctx, `
		INSERT INTO article (title, body, url, source, source_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, body, url, source, source_id, created_at, updated_at
	`, req.Title, req.Body, req.URL, req.Source, req.SourceID).Scan(
		&article.ID, &article.Title, &article.Body, &article.URL,
		&article.Source, &article.SourceID, &article.CreatedAt, &article.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(req.Categories) > 0 {
		categories, err := s.getOrCreateCategoriesTx(ctx, tx, req.Categories)
		if err != nil {
			return nil, err
		}

		for _, cat := range categories {
			_, err = tx.Exec(ctx, `
				INSERT INTO article_category (article_id, category_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, article.ID, cat.ID)
			if err != nil {
				return nil, err
			}
		}
		article.Categories = categories
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &article, nil
}

func (s *Storage) getOrCreateCategoriesTx(ctx context.Context, tx pgx.Tx, titles []string) ([]dto.Category, error) {
	categories := make([]dto.Category, 0, len(titles))
	for _, title := range titles {
		var cat dto.Category
		err := tx.QueryRow(ctx, `
			INSERT INTO category (title)
			VALUES ($1)
			ON CONFLICT (title) DO UPDATE SET title = EXCLUDED.title
			RETURNING id, title, created_at, updated_at
		`, title).Scan(&cat.ID, &cat.Title, &cat.CreatedAt, &cat.UpdatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (s *Storage) GetArticleByID(ctx context.Context, id int) (*dto.Article, error) {
	var article dto.Article
	err := s.pg.QueryRow(ctx, `
		SELECT id, title, body, url, source, source_id, created_at, updated_at
		FROM article
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&article.ID, &article.Title, &article.Body, &article.URL,
		&article.Source, &article.SourceID, &article.CreatedAt, &article.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	categories, err := s.getArticleCategories(ctx, article.ID)
	if err != nil {
		return nil, err
	}
	article.Categories = categories

	return &article, nil
}

func (s *Storage) GetArticleBySourceID(ctx context.Context, source dto.ArticleSource, sourceID string) (*dto.Article, error) {
	var article dto.Article
	err := s.pg.QueryRow(ctx, `
		SELECT id, title, body, url, source, source_id, created_at, updated_at
		FROM article
		WHERE source = $1 AND source_id = $2 AND deleted_at IS NULL
	`, source, sourceID).Scan(
		&article.ID, &article.Title, &article.Body, &article.URL,
		&article.Source, &article.SourceID, &article.CreatedAt, &article.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	categories, err := s.getArticleCategories(ctx, article.ID)
	if err != nil {
		return nil, err
	}
	article.Categories = categories

	return &article, nil
}

func (s *Storage) getArticleCategories(ctx context.Context, articleID int) ([]dto.Category, error) {
	rows, err := s.pg.Query(ctx, `
		SELECT c.id, c.title, c.created_at, c.updated_at
		FROM category c
		JOIN article_category ac ON ac.category_id = c.id
		WHERE ac.article_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.title
	`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []dto.Category
	for rows.Next() {
		var cat dto.Category
		if err := rows.Scan(&cat.ID, &cat.Title, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, rows.Err()
}

func (s *Storage) ListArticles(ctx context.Context, filter dto.ArticleFilter) ([]dto.Article, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, "a.deleted_at IS NULL")

	if filter.Source != nil {
		conditions = append(conditions, fmt.Sprintf("a.source = $%d", argIdx))
		args = append(args, *filter.Source)
		argIdx++
	}

	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`
			EXISTS (SELECT 1 FROM article_category ac WHERE ac.article_id = a.id AND ac.category_id = $%d)
		`, argIdx))
		args = append(args, *filter.CategoryID)
		argIdx++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(a.title ILIKE $%d OR a.body ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM article a WHERE %s", whereClause)
	if err := s.pg.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT a.id, a.title, a.body, a.url, a.source, a.source_id, a.created_at, a.updated_at
		FROM article a
		WHERE %s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.pg.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var articles []dto.Article
	for rows.Next() {
		var article dto.Article
		if err := rows.Scan(
			&article.ID, &article.Title, &article.Body, &article.URL,
			&article.Source, &article.SourceID, &article.CreatedAt, &article.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		articles = append(articles, article)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	for i := range articles {
		categories, err := s.getArticleCategories(ctx, articles[i].ID)
		if err != nil {
			return nil, 0, err
		}
		articles[i].Categories = categories
	}

	return articles, total, nil
}

func (s *Storage) UpdateArticle(ctx context.Context, id int, req dto.UpdateArticleRequest) (*dto.Article, error) {
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var sets []string
	var args []interface{}
	argIdx := 1

	if req.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *req.Title)
		argIdx++
	}
	if req.Body != nil {
		sets = append(sets, fmt.Sprintf("body = $%d", argIdx))
		args = append(args, *req.Body)
		argIdx++
	}
	if req.URL != nil {
		sets = append(sets, fmt.Sprintf("url = $%d", argIdx))
		args = append(args, *req.URL)
		argIdx++
	}

	sets = append(sets, "updated_at = NOW()")

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE article
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING id, title, body, url, source, source_id, created_at, updated_at
	`, strings.Join(sets, ", "), argIdx)

	var article dto.Article
	err = tx.QueryRow(ctx, query, args...).Scan(
		&article.ID, &article.Title, &article.Body, &article.URL,
		&article.Source, &article.SourceID, &article.CreatedAt, &article.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if req.Categories != nil {
		_, err = tx.Exec(ctx, "DELETE FROM article_category WHERE article_id = $1", id)
		if err != nil {
			return nil, err
		}

		if len(req.Categories) > 0 {
			categories, err := s.getOrCreateCategoriesTx(ctx, tx, req.Categories)
			if err != nil {
				return nil, err
			}

			for _, cat := range categories {
				_, err = tx.Exec(ctx, `
					INSERT INTO article_category (article_id, category_id)
					VALUES ($1, $2)
					ON CONFLICT DO NOTHING
				`, article.ID, cat.ID)
				if err != nil {
					return nil, err
				}
			}
			article.Categories = categories
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	if req.Categories == nil {
		categories, err := s.getArticleCategories(ctx, article.ID)
		if err != nil {
			return nil, err
		}
		article.Categories = categories
	}

	return &article, nil
}

func (s *Storage) DeleteArticle(ctx context.Context, id int) error {
	result, err := s.pg.Exec(ctx, `
		UPDATE article
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

func (s *Storage) ArticleExistsBySourceID(ctx context.Context, source dto.ArticleSource, sourceID string) (bool, error) {
	var exists bool
	err := s.pg.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM article
			WHERE source = $1 AND source_id = $2 AND deleted_at IS NULL
		)
	`, source, sourceID).Scan(&exists)
	return exists, err
}
