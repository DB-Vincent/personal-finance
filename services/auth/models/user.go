package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Email          string    `json:"email" db:"email"`
	PasswordHash   string    `json:"-" db:"password_hash"`
	DisplayName    string    `json:"display_name" db:"display_name"`
	CurrencySymbol string    `json:"currency_symbol" db:"currency_symbol"`
	Role           string    `json:"role" db:"role"`
	IsDisabled     bool      `json:"is_disabled" db:"is_disabled"`
	CreateTime     time.Time `json:"create_time" db:"create_time"`
	UpdateTime     time.Time `json:"update_time" db:"update_time"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateProfileRequest struct {
	DisplayName    *string `json:"display_name" validate:"omitempty,max=255"`
	CurrencySymbol *string `json:"currency_symbol" validate:"omitempty,max=10"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
