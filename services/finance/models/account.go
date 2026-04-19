package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Account struct {
	ID              uuid.UUID       `json:"id"`
	UserID          uuid.UUID       `json:"user_id"`
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	StartingBalance decimal.Decimal `json:"starting_balance"`
	Balance         decimal.Decimal `json:"balance"`
	IsArchived      bool            `json:"is_archived"`
	CreatedBy       uuid.UUID       `json:"created_by"`
	CreateTime      time.Time       `json:"create_time"`
	UpdatedBy       uuid.UUID       `json:"updated_by"`
	UpdateTime      time.Time       `json:"update_time"`
}

type CreateAccountRequest struct {
	Name            string          `json:"name" validate:"required,max=255"`
	Type            string          `json:"type" validate:"required,oneof=checking savings credit_card cash investment loan other"`
	StartingBalance decimal.Decimal `json:"starting_balance"`
}

type UpdateAccountRequest struct {
	Name *string `json:"name" validate:"omitempty,max=255"`
	Type *string `json:"type" validate:"omitempty,oneof=checking savings credit_card cash investment loan other"`
}

type NetWorth struct {
	Total decimal.Decimal `json:"total"`
}
