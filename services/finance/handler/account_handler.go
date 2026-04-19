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

type AccountHandler struct {
	svc      *service.AccountService
	validate *validator.Validate
}

func NewAccountHandler(svc *service.AccountService) *AccountHandler {
	return &AccountHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

// ListAccounts godoc
// @Summary List accounts
// @Description List all accounts for the authenticated user with computed balances
// @Tags accounts
// @Produce json
// @Param include_archived query bool false "Include archived accounts"
// @Success 200 {array} models.Account
// @Failure 401 {object} response.errorResponse
// @Router /accounts [get]
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	includeArchived := r.URL.Query().Get("include_archived") == "true"

	accounts, err := h.svc.List(r.Context(), userID, includeArchived)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}

	response.JSON(w, http.StatusOK, accounts)
}

// CreateAccount godoc
// @Summary Create an account
// @Description Create a new financial account
// @Tags accounts
// @Accept json
// @Produce json
// @Param body body models.CreateAccountRequest true "Account details"
// @Success 201 {object} models.Account
// @Failure 400 {object} response.errorResponse
// @Router /accounts [post]
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		details := validationErrors(err)
		response.Error(w, http.StatusBadRequest, "validation failed", details...)
		return
	}

	acct, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	response.JSON(w, http.StatusCreated, acct)
}

// GetAccount godoc
// @Summary Get an account
// @Description Get account details with computed balance
// @Tags accounts
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} models.Account
// @Failure 404 {object} response.errorResponse
// @Router /accounts/{id} [get]
func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	acct, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, service.ErrAccountNotFound) {
			response.Error(w, http.StatusNotFound, "account not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to get account")
		return
	}

	response.JSON(w, http.StatusOK, acct)
}

// UpdateAccount godoc
// @Summary Update an account
// @Description Update account name or type
// @Tags accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Param body body models.UpdateAccountRequest true "Fields to update"
// @Success 200 {object} models.Account
// @Failure 400 {object} response.errorResponse
// @Failure 404 {object} response.errorResponse
// @Router /accounts/{id} [put]
func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	var req models.UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		details := validationErrors(err)
		response.Error(w, http.StatusBadRequest, "validation failed", details...)
		return
	}

	acct, err := h.svc.Update(r.Context(), userID, id, req)
	if err != nil {
		if errors.Is(err, service.ErrAccountNotFound) {
			response.Error(w, http.StatusNotFound, "account not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update account")
		return
	}

	response.JSON(w, http.StatusOK, acct)
}

// ArchiveAccount godoc
// @Summary Toggle account archive status
// @Description Archive or unarchive an account
// @Tags accounts
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} models.Account
// @Failure 404 {object} response.errorResponse
// @Router /accounts/{id}/archive [post]
func (h *AccountHandler) ToggleArchive(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	acct, err := h.svc.ToggleArchive(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, service.ErrAccountNotFound) {
			response.Error(w, http.StatusNotFound, "account not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to toggle archive")
		return
	}

	response.JSON(w, http.StatusOK, acct)
}

// DeleteAccount godoc
// @Summary Delete an account
// @Description Delete an account (fails if it has transactions)
// @Tags accounts
// @Produce json
// @Param id path string true "Account ID"
// @Success 204
// @Failure 404 {object} response.errorResponse
// @Failure 409 {object} response.errorResponse
// @Router /accounts/{id} [delete]
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		if errors.Is(err, service.ErrAccountNotFound) {
			response.Error(w, http.StatusNotFound, "account not found")
			return
		}
		if errors.Is(err, service.ErrAccountHasTransactions) {
			response.Error(w, http.StatusConflict, "account has transactions and cannot be deleted")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to delete account")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// NetWorth godoc
// @Summary Get net worth
// @Description Sum of all non-archived account balances
// @Tags accounts
// @Produce json
// @Success 200 {object} models.NetWorth
// @Failure 401 {object} response.errorResponse
// @Router /accounts/net-worth [get]
func (h *AccountHandler) NetWorth(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromHeader(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "missing or invalid user ID")
		return
	}

	nw, err := h.svc.NetWorth(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to calculate net worth")
		return
	}

	response.JSON(w, http.StatusOK, nw)
}
