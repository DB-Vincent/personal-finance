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

type TagHandler struct {
	svc      *service.TagService
	validate *validator.Validate
}

func NewTagHandler(svc *service.TagService) *TagHandler {
	return &TagHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

// ListTags godoc
// @Summary List tags
// @Description List all tags for the authenticated user
// @Tags tags
// @Produce json
// @Success 200 {array} models.Tag
// @Failure 401 {object} response.errorResponse
// @Router /tags [get]
func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	tags, err := h.svc.List(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list tags")
		return
	}

	response.JSON(w, http.StatusOK, tags)
}

// CreateTag godoc
// @Summary Create a tag
// @Description Create a new tag with name and color
// @Tags tags
// @Accept json
// @Produce json
// @Param body body models.CreateTagRequest true "Tag details"
// @Success 201 {object} models.Tag
// @Failure 400 {object} response.errorResponse
// @Failure 409 {object} response.errorResponse
// @Router /tags [post]
func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	var req models.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		details := validationErrors(err)
		response.Error(w, http.StatusBadRequest, "validation failed", details...)
		return
	}

	tag, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrTagExists) {
			response.Error(w, http.StatusConflict, "tag already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to create tag")
		return
	}

	response.JSON(w, http.StatusCreated, tag)
}

// UpdateTag godoc
// @Summary Update a tag
// @Description Update tag name or color
// @Tags tags
// @Accept json
// @Produce json
// @Param id path string true "Tag ID"
// @Param body body models.UpdateTagRequest true "Fields to update"
// @Success 200 {object} models.Tag
// @Failure 400 {object} response.errorResponse
// @Failure 404 {object} response.errorResponse
// @Router /tags/{id} [put]
func (h *TagHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid tag ID")
		return
	}

	var req models.UpdateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		details := validationErrors(err)
		response.Error(w, http.StatusBadRequest, "validation failed", details...)
		return
	}

	tag, err := h.svc.Update(r.Context(), userID, id, req)
	if err != nil {
		if errors.Is(err, service.ErrTagNotFound) {
			response.Error(w, http.StatusNotFound, "tag not found")
			return
		}
		if errors.Is(err, service.ErrTagExists) {
			response.Error(w, http.StatusConflict, "tag name already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update tag")
		return
	}

	response.JSON(w, http.StatusOK, tag)
}

// DeleteTag godoc
// @Summary Delete a tag
// @Description Delete a tag (removes from all transactions)
// @Tags tags
// @Produce json
// @Param id path string true "Tag ID"
// @Success 204
// @Failure 404 {object} response.errorResponse
// @Router /tags/{id} [delete]
func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid tag ID")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		if errors.Is(err, service.ErrTagNotFound) {
			response.Error(w, http.StatusNotFound, "tag not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to delete tag")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
