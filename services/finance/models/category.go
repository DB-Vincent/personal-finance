package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	GroupName  string    `json:"group_name"`
	Name       string    `json:"name"`
	IsIncome   bool      `json:"is_income"`
	IsArchived bool      `json:"is_archived"`
	CreateTime time.Time `json:"create_time"`
}

type CreateCategoryRequest struct {
	GroupName string `json:"group_name" validate:"required,max=100"`
	Name      string `json:"name" validate:"required,max=100"`
	IsIncome  bool   `json:"is_income"`
}

type UpdateCategoryRequest struct {
	GroupName *string `json:"group_name" validate:"omitempty,max=100"`
	Name      *string `json:"name" validate:"omitempty,max=100"`
}

type CategoryGroup struct {
	GroupName  string     `json:"group_name"`
	Categories []Category `json:"categories"`
}
