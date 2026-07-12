package dto

import "time"

type ArticleSource string

const (
	SourceManual   ArticleSource = "manual"
	SourceVK       ArticleSource = "vk"
	SourceTelegram ArticleSource = "telegram"
	SourceHabr     ArticleSource = "habr"
)

type Article struct {
	ID         int           `json:"id"`
	Title      string        `json:"title"`
	Body       string        `json:"body"`
	URL        string        `json:"url"`
	Source     ArticleSource `json:"source"`
	SourceID   *string       `json:"source_id,omitempty"`
	Categories []Category    `json:"categories"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  *time.Time    `json:"updated_at,omitempty"`
}

type CreateArticleRequest struct {
	Title      string        `json:"title"`
	Body       string        `json:"body"`
	URL        string        `json:"url"`
	Source     ArticleSource `json:"source"`
	SourceID   *string       `json:"source_id,omitempty"`
	Categories []string      `json:"categories"`
}

type UpdateArticleRequest struct {
	Title      *string  `json:"title,omitempty"`
	Body       *string  `json:"body,omitempty"`
	URL        *string  `json:"url,omitempty"`
	Categories []string `json:"categories,omitempty"`
}

type ArticleFilter struct {
	Source     *ArticleSource
	CategoryID *int
	Search     *string
	Limit      int
	Offset     int
}
