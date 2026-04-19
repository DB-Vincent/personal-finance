package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DB-Vincent/personal-finance/pkg/response"
	"github.com/DB-Vincent/personal-finance/services/auth/models"
	"github.com/DB-Vincent/personal-finance/services/auth/service"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	auth     *service.AuthService
	validate *validator.Validate
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{
		auth:     auth,
		validate: validator.New(),
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account and return auth tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param body body models.RegisterRequest true "Registration details"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} response.errorResponse
// @Failure 403 {object} response.errorResponse
// @Failure 409 {object} response.errorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		details := validationErrors(err)
		response.Error(w, http.StatusBadRequest, "validation failed", details...)
		return
	}

	resp, err := h.auth.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRegistrationDisabled):
			response.Error(w, http.StatusForbidden, "registration is disabled")
		case errors.Is(err, service.ErrEmailExists):
			response.Error(w, http.StatusConflict, "email already exists")
		default:
			response.Error(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

// Login godoc
// @Summary Log in
// @Description Authenticate with email and password, returns auth tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param body body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} response.errorResponse
// @Failure 401 {object} response.errorResponse
// @Failure 403 {object} response.errorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		details := validationErrors(err)
		response.Error(w, http.StatusBadRequest, "validation failed", details...)
		return
	}

	resp, err := h.auth.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			response.Error(w, http.StatusUnauthorized, "invalid email or password")
		case errors.Is(err, service.ErrAccountDisabled):
			response.Error(w, http.StatusForbidden, "account is disabled")
		default:
			response.Error(w, http.StatusInternalServerError, "login failed")
		}
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// Refresh godoc
// @Summary Refresh tokens
// @Description Exchange a valid refresh token for new access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param body body models.RefreshRequest true "Refresh token"
// @Success 200 {object} models.TokenResponse
// @Failure 400 {object} response.errorResponse
// @Failure 401 {object} response.errorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		details := validationErrors(err)
		response.Error(w, http.StatusBadRequest, "validation failed", details...)
		return
	}

	resp, err := h.auth.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			response.Error(w, http.StatusUnauthorized, "invalid refresh token")
		case errors.Is(err, service.ErrAccountDisabled):
			response.Error(w, http.StatusForbidden, "account is disabled")
		default:
			response.Error(w, http.StatusInternalServerError, "token refresh failed")
		}
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func validationErrors(err error) []response.ErrorDetail {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}
	details := make([]response.ErrorDetail, len(ve))
	for i, fe := range ve {
		details[i] = response.ErrorDetail{
			Field:  fe.Field(),
			Reason: fe.Tag(),
		}
	}
	return details
}
