package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID                uuid.UUID       `json:"id"`
	UserID            uuid.UUID       `json:"user_id"`
	AccountID         uuid.UUID       `json:"account_id"`
	Type              string          `json:"type"`
	Amount            decimal.Decimal `json:"amount"`
	CategoryID        *uuid.UUID      `json:"category_id,omitempty"`
	TransferAccountID *uuid.UUID      `json:"transfer_account_id,omitempty"`
	Date              time.Time       `json:"date"`
	Notes             *string         `json:"notes,omitempty"`
	RecurringRuleID   *uuid.UUID      `json:"recurring_rule_id,omitempty"`
	Tags              []Tag           `json:"tags,omitempty"`
	CreatedBy         uuid.UUID       `json:"created_by"`
	CreateTime        time.Time       `json:"create_time"`
	UpdatedBy         uuid.UUID       `json:"updated_by"`
	UpdateTime        time.Time       `json:"update_time"`
}

type CreateTransactionRequest struct {
	AccountID         uuid.UUID       `json:"account_id" validate:"required"`
	Type              string          `json:"type" validate:"required,oneof=income expense transfer"`
	Amount            decimal.Decimal `json:"amount" validate:"required"`
	CategoryID        *uuid.UUID      `json:"category_id"`
	TransferAccountID *uuid.UUID      `json:"transfer_account_id"`
	Date              string          `json:"date" validate:"required"`
	Notes             *string         `json:"notes"`
	TagIDs            []uuid.UUID     `json:"tag_ids"`
}

type UpdateTransactionRequest struct {
	AccountID         *uuid.UUID       `json:"account_id"`
	Type              *string          `json:"type" validate:"omitempty,oneof=income expense transfer"`
	Amount            *decimal.Decimal `json:"amount"`
	CategoryID        *uuid.UUID       `json:"category_id"`
	TransferAccountID *uuid.UUID       `json:"transfer_account_id"`
	Date              *string          `json:"date"`
	Notes             *string          `json:"notes"`
	TagIDs            *[]uuid.UUID     `json:"tag_ids"`
}

type TransactionFilter struct {
	AccountID  *uuid.UUID
	CategoryID *uuid.UUID
	TagID      *uuid.UUID
	Type       *string
	DateFrom   *time.Time
	DateTo     *time.Time
	AmountMin  *decimal.Decimal
	AmountMax  *decimal.Decimal
	Search     *string
	PageSize   int
	PageToken  string
}
