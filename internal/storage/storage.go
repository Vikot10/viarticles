package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("article not found")

type Article struct {
	ID        int64
	Title     string
	Body      string
	URL       string
	Source    string
	Read      bool
	Liked     bool
	CreatedAt string
}

type NewArticle struct {
	Title    string
	Body     string
	URL      string
	Source   string
	SourceID *string
}

type Storage struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*Storage, error) {
	if path != ":memory:" {
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolve database path: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
		path = filepath.ToSlash(absolutePath)
	}

	dsn := "file:" + path + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	store := &Storage{db: db}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if err := store.migrate(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) migrate(ctx context.Context) error {
	const schema = `
		CREATE TABLE IF NOT EXISTS article (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			body TEXT NOT NULL DEFAULT '',
			url TEXT NOT NULL,
			source TEXT NOT NULL CHECK (source IN ('manual', 'telegram')),
			source_id TEXT,
			is_read INTEGER NOT NULL DEFAULT 0,
			is_liked INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT,
			deleted_at TEXT,
			UNIQUE (source, source_id)
		);
		CREATE INDEX IF NOT EXISTS article_active_created_idx
			ON article (created_at DESC) WHERE deleted_at IS NULL;
		CREATE TABLE IF NOT EXISTS app_state (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate sqlite: %w", err)
	}
	return nil
}

func (s *Storage) CreateArticle(ctx context.Context, article NewArticle) (Article, error) {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO article (title, body, url, source, source_id)
		VALUES (?, ?, ?, ?, ?)
	`, article.Title, article.Body, article.URL, article.Source, article.SourceID)
	if err != nil {
		return Article{}, fmt.Errorf("create article: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Article{}, fmt.Errorf("read article id: %w", err)
	}
	return s.GetArticle(ctx, id)
}

func (s *Storage) CreateImportedArticle(ctx context.Context, article NewArticle) (bool, error) {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO article (title, body, url, source, source_id)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (source, source_id) DO NOTHING
	`, article.Title, article.Body, article.URL, article.Source, article.SourceID)
	if err != nil {
		return false, fmt.Errorf("create imported article: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read imported result: %w", err)
	}
	return rows > 0, nil
}

func (s *Storage) GetArticle(ctx context.Context, id int64) (Article, error) {
	var article Article
	err := s.db.QueryRowContext(ctx, `
		SELECT id, title, body, url, source, is_read, is_liked,
		       strftime('%Y-%m-%d %H:%M', created_at)
		FROM article
		WHERE id = ? AND deleted_at IS NULL
	`, id).Scan(
		&article.ID, &article.Title, &article.Body, &article.URL, &article.Source,
		&article.Read, &article.Liked, &article.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Article{}, ErrNotFound
	}
	if err != nil {
		return Article{}, fmt.Errorf("get article: %w", err)
	}
	return article, nil
}

func (s *Storage) ListArticles(ctx context.Context, filter, search string) ([]Article, error) {
	conditions := []string{"deleted_at IS NULL"}
	args := make([]any, 0, 1)
	switch filter {
	case "unread":
		conditions = append(conditions, "is_read = 0")
	case "liked":
		conditions = append(conditions, "is_liked = 1")
	case "telegram":
		conditions = append(conditions, "source = 'telegram'")
	case "manual":
		conditions = append(conditions, "source = 'manual'")
	}
	if search != "" {
		conditions = append(conditions, "(title LIKE ? OR body LIKE ? OR url LIKE ?)")
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern, pattern)
	}

	query := `SELECT id, title, body, url, source, is_read, is_liked,
		strftime('%Y-%m-%d %H:%M', created_at)
		FROM article WHERE ` + strings.Join(conditions, " AND ") + `
		ORDER BY created_at DESC, id DESC LIMIT 200`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}
	defer rows.Close()

	articles := make([]Article, 0)
	for rows.Next() {
		var article Article
		if err := rows.Scan(
			&article.ID, &article.Title, &article.Body, &article.URL, &article.Source,
			&article.Read, &article.Liked, &article.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}
		articles = append(articles, article)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate articles: %w", err)
	}
	return articles, nil
}

func (s *Storage) ToggleRead(ctx context.Context, id int64) (Article, error) {
	return s.toggle(ctx, id, "is_read")
}

func (s *Storage) ToggleLiked(ctx context.Context, id int64) (Article, error) {
	return s.toggle(ctx, id, "is_liked")
}

func (s *Storage) toggle(ctx context.Context, id int64, column string) (Article, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE article SET `+column+` = NOT `+column+`,
		updated_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL`, id)
	if err != nil {
		return Article{}, fmt.Errorf("toggle %s: %w", column, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return Article{}, err
	}
	if rows == 0 {
		return Article{}, ErrNotFound
	}
	return s.GetArticle(ctx, id)
}

func (s *Storage) DeleteArticle(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `UPDATE article SET deleted_at = CURRENT_TIMESTAMP,
		updated_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("delete article: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Storage) State(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read state: %w", err)
	}
	return value, nil
}

func (s *Storage) SetState(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO app_state (key, value) VALUES (?, ?)
		ON CONFLICT (key) DO UPDATE SET value = excluded.value`, key, value)
	if err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}
