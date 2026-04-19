package service

import (
	"context"
	"errors"

	"github.com/DB-Vincent/personal-finance/services/finance/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrAccountNotFound       = errors.New("account not found")
	ErrAccountHasTransactions = errors.New("account has transactions")
)

type AccountRepository interface {
	Create(ctx context.Context, acct *models.Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
	ListByUser(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]models.Account, error)
	Update(ctx context.Context, acct *models.Account) error
	SetArchived(ctx context.Context, id uuid.UUID, archived bool) error
	Delete(ctx context.Context, id uuid.UUID) error
	HasTransactions(ctx context.Context, id uuid.UUID) (bool, error)
	NetWorth(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
}

type AccountService struct {
	repo AccountRepository
}

func NewAccountService(repo AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) Create(ctx context.Context, userID uuid.UUID, req models.CreateAccountRequest) (*models.Account, error) {
	acct := &models.Account{
		UserID:          userID,
		Name:            req.Name,
		Type:            req.Type,
		StartingBalance: req.StartingBalance,
		Balance:         req.StartingBalance,
		CreatedBy:       userID,
		UpdatedBy:       userID,
	}

	if err := s.repo.Create(ctx, acct); err != nil {
		return nil, err
	}

	return acct, nil
}

func (s *AccountService) Get(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*models.Account, error) {
	acct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrAccountNotFound
	}
	if acct.UserID != userID {
		return nil, ErrAccountNotFound
	}
	return acct, nil
}

func (s *AccountService) List(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]models.Account, error) {
	return s.repo.ListByUser(ctx, userID, includeArchived)
}

func (s *AccountService) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, req models.UpdateAccountRequest) (*models.Account, error) {
	acct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrAccountNotFound
	}
	if acct.UserID != userID {
		return nil, ErrAccountNotFound
	}

	if req.Name != nil {
		acct.Name = *req.Name
	}
	if req.Type != nil {
		acct.Type = *req.Type
	}
	acct.UpdatedBy = userID

	if err := s.repo.Update(ctx, acct); err != nil {
		return nil, err
	}

	return acct, nil
}

func (s *AccountService) ToggleArchive(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*models.Account, error) {
	acct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrAccountNotFound
	}
	if acct.UserID != userID {
		return nil, ErrAccountNotFound
	}

	if err := s.repo.SetArchived(ctx, id, !acct.IsArchived); err != nil {
		return nil, err
	}

	acct.IsArchived = !acct.IsArchived
	return acct, nil
}

func (s *AccountService) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	acct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrAccountNotFound
	}
	if acct.UserID != userID {
		return ErrAccountNotFound
	}

	hasTx, err := s.repo.HasTransactions(ctx, id)
	if err != nil {
		return err
	}
	if hasTx {
		return ErrAccountHasTransactions
	}

	return s.repo.Delete(ctx, id)
}

func (s *AccountService) NetWorth(ctx context.Context, userID uuid.UUID) (*models.NetWorth, error) {
	total, err := s.repo.NetWorth(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &models.NetWorth{Total: total}, nil
}
