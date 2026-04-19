package models

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Name       string    `json:"name"`
	Color      string    `json:"color"`
	CreateTime time.Time `json:"create_time"`
}

type CreateTagRequest struct {
	Name  string `json:"name" validate:"required,max=100"`
	Color string `json:"color" validate:"omitempty,max=7"`
}

type UpdateTagRequest struct {
	Name  *string `json:"name" validate:"omitempty,max=100"`
	Color *string `json:"color" validate:"omitempty,max=7"`
}
