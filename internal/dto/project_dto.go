package dto

import "encoding/json"

type CreateProjectInput struct {
	Title       string          `json:"title" binding:"required"`
	Description string          `json:"description"`
	Content     json.RawMessage `json:"content"  binding:"required"`
}

type UpdateProjectInput struct {
	Title       *string          `json:"title"`
	Description *string          `json:"description"`
	Content     *json.RawMessage `json:"content"`
}
