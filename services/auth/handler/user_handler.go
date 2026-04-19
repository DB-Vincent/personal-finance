package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DB-Vincent/personal-finance/pkg/response"
	"github.com/DB-Vincent/personal-finance/services/auth/models"
	"github.com/DB-Vincent/personal-finance/services/auth/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type UserHandler struct {
	auth     *service.AuthService
	validate *validator.Validate
}

func NewUserHandler(auth *service.AuthService) *UserHandler {
	return &UserHandler{
		auth:     auth,
		validate: validator.New(),
	}
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Returns the authenticated user's profile
// @Tags users
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} response.errorResponse
// @Failure 404 {object} response.errorResponse
// @Router /users/me [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	user, err := h.auth.GetProfile(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}

	response.JSON(w, http.StatusOK, user)
}

// UpdateProfile godoc
// @Summary Update current user profile
// @Description Update display name and/or currency symbol for the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Param body body models.UpdateProfileRequest true "Profile fields to update"
// @Success 200 {object} models.User
// @Failure 400 {object} response.errorResponse
// @Failure 401 {object} response.errorResponse
// @Router /users/me [put]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			details := make([]response.ErrorDetail, len(ve))
			for i, fe := range ve {
				details[i] = response.ErrorDetail{
					Field:  fe.Field(),
					Reason: fe.Tag(),
				}
			}
			response.Error(w, http.StatusBadRequest, "validation failed", details...)
			return
		}
		response.Error(w, http.StatusBadRequest, "validation failed")
		return
	}

	user, err := h.auth.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	response.JSON(w, http.StatusOK, user)
}

func userIDFromHeader(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(r.Header.Get("X-User-ID"))
}
