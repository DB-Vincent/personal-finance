package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/DB-Vincent/personal-finance/pkg/response"
	"github.com/DB-Vincent/personal-finance/services/finance/models"
	"github.com/DB-Vincent/personal-finance/services/finance/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionHandler struct {
	svc      *service.TransactionService
	validate *validator.Validate
}

func NewTransactionHandler(svc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

// ListTransactions godoc
// @Summary List transactions
// @Description List transactions with cursor-based pagination and filters
// @Tags transactions
// @Produce json
// @Param page_size query int false "Page size (default 50, max 100)"
// @Param page_token query string false "Cursor for next page"
// @Param account_id query string false "Filter by account ID"
// @Param category_id query string false "Filter by category ID"
// @Param tag_id query string false "Filter by tag ID"
// @Param type query string false "Filter by type (income, expense, transfer)"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Param amount_min query number false "Minimum amount"
// @Param amount_max query number false "Maximum amount"
// @Param search query string false "Search in notes"
// @Success 200 {object} response.ListResponse
// @Failure 401 {object} response.errorResponse
// @Router /transactions [get]
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	filter := models.TransactionFilter{
		PageToken: r.URL.Query().Get("page_token"),
	}

	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil {
			filter.PageSize = v
		}
	}

	if v := r.URL.Query().Get("account_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.AccountID = &id
		}
	}
	if v := r.URL.Query().Get("category_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.CategoryID = &id
		}
	}
	if v := r.URL.Query().Get("tag_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.TagID = &id
		}
	}
	if v := r.URL.Query().Get("type"); v != "" {
		filter.Type = &v
	}
	if v := r.URL.Query().Get("date_from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.DateFrom = &t
		}
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.DateTo = &t
		}
	}
	if v := r.URL.Query().Get("amount_min"); v != "" {
		if d, err := decimal.NewFromString(v); err == nil {
			filter.AmountMin = &d
		}
	}
	if v := r.URL.Query().Get("amount_max"); v != "" {
		if d, err := decimal.NewFromString(v); err == nil {
			filter.AmountMax = &d
		}
	}
	if v := r.URL.Query().Get("search"); v != "" {
		filter.Search = &v
	}

	txs, nextToken, total, err := h.svc.List(r.Context(), userID, filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}

	response.List(w, txs, nextToken, total)
}

// CreateTransaction godoc
// @Summary Create a transaction
// @Description Create a new transaction (income, expense, or transfer)
// @Tags transactions
// @Accept json
// @Produce json
// @Param body body models.CreateTransactionRequest true "Transaction details"
// @Success 201 {object} models.Transaction
// @Failure 400 {object} response.errorResponse
// @Router /transactions [post]
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	var req models.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		details := validationErrors(err)
		response.Error(w, http.StatusBadRequest, "validation failed", details...)
		return
	}

	tx, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAmount):
			response.Error(w, http.StatusBadRequest, "amount must be positive")
		case errors.Is(err, service.ErrTransferRequiresTarget):
			response.Error(w, http.StatusBadRequest, "transfer requires a destination account")
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.JSON(w, http.StatusCreated, tx)
}

// GetTransaction godoc
// @Summary Get a transaction
// @Description Get a single transaction by ID
// @Tags transactions
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} models.Transaction
// @Failure 404 {object} response.errorResponse
// @Router /transactions/{id} [get]
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid transaction ID")
		return
	}

	tx, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, service.ErrTransactionNotFound) {
			response.Error(w, http.StatusNotFound, "transaction not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to get transaction")
		return
	}

	response.JSON(w, http.StatusOK, tx)
}

// UpdateTransaction godoc
// @Summary Update a transaction
// @Description Update transaction fields
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Param body body models.UpdateTransactionRequest true "Fields to update"
// @Success 200 {object} models.Transaction
// @Failure 400 {object} response.errorResponse
// @Failure 404 {object} response.errorResponse
// @Router /transactions/{id} [put]
func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid transaction ID")
		return
	}

	var req models.UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		details := validationErrors(err)
		response.Error(w, http.StatusBadRequest, "validation failed", details...)
		return
	}

	tx, err := h.svc.Update(r.Context(), userID, id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTransactionNotFound):
			response.Error(w, http.StatusNotFound, "transaction not found")
		case errors.Is(err, service.ErrInvalidAmount):
			response.Error(w, http.StatusBadRequest, "amount must be positive")
		case errors.Is(err, service.ErrTransferRequiresTarget):
			response.Error(w, http.StatusBadRequest, "transfer requires a destination account")
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.JSON(w, http.StatusOK, tx)
}

// DeleteTransaction godoc
// @Summary Delete a transaction
// @Description Permanently delete a transaction
// @Tags transactions
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 204
// @Failure 404 {object} response.errorResponse
// @Router /transactions/{id} [delete]
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid transaction ID")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		if errors.Is(err, service.ErrTransactionNotFound) {
			response.Error(w, http.StatusNotFound, "transaction not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to delete transaction")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
