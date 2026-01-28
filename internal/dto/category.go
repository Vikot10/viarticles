package dto

import "time"

type Category struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type CreateCategoryRequest struct {
	Title string `json:"title"`
}

type UpdateCategoryRequest struct {
	Title string `json:"title"`
}
