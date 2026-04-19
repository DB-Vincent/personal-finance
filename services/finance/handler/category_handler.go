package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DB-Vincent/personal-finance/pkg/response"
	"github.com/DB-Vincent/personal-finance/services/finance/models"
	"github.com/DB-Vincent/personal-finance/services/finance/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	svc      *service.CategoryService
	validate *validator.Validate
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

// ListCategories godoc
// @Summary List categories
// @Description List all categories for the authenticated user, grouped by group_name
// @Tags categories
// @Produce json
// @Param include_archived query bool false "Include archived categories"
// @Success 200 {array} models.CategoryGroup
// @Failure 401 {object} response.errorResponse
// @Router /categories [get]
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	includeArchived := r.URL.Query().Get("include_archived") == "true"

	groups, err := h.svc.List(r.Context(), userID, includeArchived)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list categories")
		return
	}

	response.JSON(w, http.StatusOK, groups)
}

// CreateCategory godoc
// @Summary Create a category
// @Description Create a new custom category
// @Tags categories
// @Accept json
// @Produce json
// @Param body body models.CreateCategoryRequest true "Category details"
// @Success 201 {object} models.Category
// @Failure 400 {object} response.errorResponse
// @Failure 409 {object} response.errorResponse
// @Router /categories [post]
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	var req models.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		details := validationErrors(err)
		response.Error(w, http.StatusBadRequest, "validation failed", details...)
		return
	}

	cat, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrCategoryExists) {
			response.Error(w, http.StatusConflict, "category already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	response.JSON(w, http.StatusCreated, cat)
}

// UpdateCategory godoc
// @Summary Update a category
// @Description Rename a category or change its group
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param body body models.UpdateCategoryRequest true "Fields to update"
// @Success 200 {object} models.Category
// @Failure 400 {object} response.errorResponse
// @Failure 404 {object} response.errorResponse
// @Router /categories/{id} [put]
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	var req models.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		details := validationErrors(err)
		response.Error(w, http.StatusBadRequest, "validation failed", details...)
		return
	}

	cat, err := h.svc.Update(r.Context(), userID, id, req)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			response.Error(w, http.StatusNotFound, "category not found")
			return
		}
		if errors.Is(err, service.ErrCategoryExists) {
			response.Error(w, http.StatusConflict, "category name already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update category")
		return
	}

	response.JSON(w, http.StatusOK, cat)
}

// ArchiveCategory godoc
// @Summary Toggle category archive status
// @Description Archive or unarchive a category
// @Tags categories
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} models.Category
// @Failure 404 {object} response.errorResponse
// @Router /categories/{id}/archive [post]
func (h *CategoryHandler) ToggleArchive(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	cat, err := h.svc.ToggleArchive(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			response.Error(w, http.StatusNotFound, "category not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to toggle archive")
		return
	}

	response.JSON(w, http.StatusOK, cat)
}
