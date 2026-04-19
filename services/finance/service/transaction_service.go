package service

import (
	"context"
	"errors"
	"time"

	"github.com/DB-Vincent/personal-finance/services/finance/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrInvalidAmount          = errors.New("amount must be positive")
	ErrTransferRequiresTarget = errors.New("transfer requires a destination account")
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *models.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	List(ctx context.Context, userID uuid.UUID, filter models.TransactionFilter) ([]models.Transaction, string, int64, error)
	Update(ctx context.Context, tx *models.Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TransactionTagRepository interface {
	SetTags(ctx context.Context, transactionID uuid.UUID, tagIDs []uuid.UUID) error
	GetTags(ctx context.Context, transactionID uuid.UUID) ([]models.Tag, error)
}

type TransactionService struct {
	repo    TransactionRepository
	tagRepo TransactionTagRepository
}

func NewTransactionService(repo TransactionRepository, tagRepo TransactionTagRepository) *TransactionService {
	return &TransactionService{repo: repo, tagRepo: tagRepo}
}

func (s *TransactionService) Create(ctx context.Context, userID uuid.UUID, req models.CreateTransactionRequest) (*models.Transaction, error) {
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	if req.Type == "transfer" && req.TransferAccountID == nil {
		return nil, ErrTransferRequiresTarget
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format, expected YYYY-MM-DD")
	}

	tx := &models.Transaction{
		UserID:            userID,
		AccountID:         req.AccountID,
		Type:              req.Type,
		Amount:            req.Amount,
		CategoryID:        req.CategoryID,
		TransferAccountID: req.TransferAccountID,
		Date:              date,
		Notes:             req.Notes,
		CreatedBy:         userID,
		UpdatedBy:         userID,
	}

	if err := s.repo.Create(ctx, tx); err != nil {
		return nil, err
	}

	if len(req.TagIDs) > 0 {
		if err := s.tagRepo.SetTags(ctx, tx.ID, req.TagIDs); err != nil {
			return nil, err
		}
		tags, _ := s.tagRepo.GetTags(ctx, tx.ID)
		tx.Tags = tags
	}

	return tx, nil
}

func (s *TransactionService) Get(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*models.Transaction, error) {
	tx, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTransactionNotFound
	}
	if tx.UserID != userID {
		return nil, ErrTransactionNotFound
	}

	tags, _ := s.tagRepo.GetTags(ctx, id)
	tx.Tags = tags

	return tx, nil
}

func (s *TransactionService) List(ctx context.Context, userID uuid.UUID, filter models.TransactionFilter) ([]models.Transaction, string, int64, error) {
	txs, nextToken, total, err := s.repo.List(ctx, userID, filter)
	if err != nil {
		return nil, "", 0, err
	}

	for i := range txs {
		tags, _ := s.tagRepo.GetTags(ctx, txs[i].ID)
		txs[i].Tags = tags
	}

	return txs, nextToken, total, nil
}

func (s *TransactionService) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, req models.UpdateTransactionRequest) (*models.Transaction, error) {
	tx, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTransactionNotFound
	}
	if tx.UserID != userID {
		return nil, ErrTransactionNotFound
	}

	if req.AccountID != nil {
		tx.AccountID = *req.AccountID
	}
	if req.Type != nil {
		tx.Type = *req.Type
	}
	if req.Amount != nil {
		if req.Amount.LessThanOrEqual(decimal.Zero) {
			return nil, ErrInvalidAmount
		}
		tx.Amount = *req.Amount
	}
	if req.CategoryID != nil {
		tx.CategoryID = req.CategoryID
	}
	if req.TransferAccountID != nil {
		tx.TransferAccountID = req.TransferAccountID
	}
	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return nil, errors.New("invalid date format, expected YYYY-MM-DD")
		}
		tx.Date = date
	}
	if req.Notes != nil {
		tx.Notes = req.Notes
	}
	tx.UpdatedBy = userID

	if tx.Type == "transfer" && tx.TransferAccountID == nil {
		return nil, ErrTransferRequiresTarget
	}

	if err := s.repo.Update(ctx, tx); err != nil {
		return nil, err
	}

	if req.TagIDs != nil {
		if err := s.tagRepo.SetTags(ctx, id, *req.TagIDs); err != nil {
			return nil, err
		}
	}

	tags, _ := s.tagRepo.GetTags(ctx, id)
	tx.Tags = tags

	return tx, nil
}

func (s *TransactionService) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	tx, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrTransactionNotFound
	}
	if tx.UserID != userID {
		return ErrTransactionNotFound
	}

	return s.repo.Delete(ctx, id)
}
