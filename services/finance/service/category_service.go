package service

import (
	"context"
	"errors"
	"sync"

	"github.com/DB-Vincent/personal-finance/services/finance/models"
	"github.com/DB-Vincent/personal-finance/services/finance/seed"
	"github.com/google/uuid"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryExists   = errors.New("category already exists")
	ErrCategoryInUse    = errors.New("category is in use by transactions")
)

type CategoryRepository interface {
	Create(ctx context.Context, cat *models.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error)
	ListByUser(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]models.Category, error)
	Update(ctx context.Context, cat *models.Category) error
	SetArchived(ctx context.Context, id uuid.UUID, archived bool) error
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}

type CategoryService struct {
	repo    CategoryRepository
	seeded  map[uuid.UUID]bool
	seedMu  sync.Mutex
}

func NewCategoryService(repo CategoryRepository) *CategoryService {
	return &CategoryService{
		repo:   repo,
		seeded: make(map[uuid.UUID]bool),
	}
}

func (s *CategoryService) EnsureDefaults(ctx context.Context, userID uuid.UUID) error {
	s.seedMu.Lock()
	if s.seeded[userID] {
		s.seedMu.Unlock()
		return nil
	}
	s.seedMu.Unlock()

	count, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return err
	}

	if count == 0 {
		if err := seed.CategoriesForUser(ctx, s.repo, userID); err != nil {
			return err
		}
	}

	s.seedMu.Lock()
	s.seeded[userID] = true
	s.seedMu.Unlock()
	return nil
}

func (s *CategoryService) List(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]models.CategoryGroup, error) {
	if err := s.EnsureDefaults(ctx, userID); err != nil {
		return nil, err
	}

	cats, err := s.repo.ListByUser(ctx, userID, includeArchived)
	if err != nil {
		return nil, err
	}

	groupMap := make(map[string][]models.Category)
	var groupOrder []string
	for _, c := range cats {
		if _, exists := groupMap[c.GroupName]; !exists {
			groupOrder = append(groupOrder, c.GroupName)
		}
		groupMap[c.GroupName] = append(groupMap[c.GroupName], c)
	}

	groups := make([]models.CategoryGroup, len(groupOrder))
	for i, name := range groupOrder {
		groups[i] = models.CategoryGroup{
			GroupName:  name,
			Categories: groupMap[name],
		}
	}

	return groups, nil
}

func (s *CategoryService) Create(ctx context.Context, userID uuid.UUID, req models.CreateCategoryRequest) (*models.Category, error) {
	cat := &models.Category{
		UserID:    userID,
		GroupName: req.GroupName,
		Name:      req.Name,
		IsIncome:  req.IsIncome,
	}

	if err := s.repo.Create(ctx, cat); err != nil {
		return nil, err
	}

	return cat, nil
}

func (s *CategoryService) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, req models.UpdateCategoryRequest) (*models.Category, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCategoryNotFound
	}
	if cat.UserID != userID {
		return nil, ErrCategoryNotFound
	}

	if req.GroupName != nil {
		cat.GroupName = *req.GroupName
	}
	if req.Name != nil {
		cat.Name = *req.Name
	}

	if err := s.repo.Update(ctx, cat); err != nil {
		return nil, err
	}

	return cat, nil
}

func (s *CategoryService) ToggleArchive(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*models.Category, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCategoryNotFound
	}
	if cat.UserID != userID {
		return nil, ErrCategoryNotFound
	}

	if err := s.repo.SetArchived(ctx, id, !cat.IsArchived); err != nil {
		return nil, err
	}

	cat.IsArchived = !cat.IsArchived
	return cat, nil
}
