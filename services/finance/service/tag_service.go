package service

import (
	"context"
	"errors"

	"github.com/DB-Vincent/personal-finance/services/finance/models"
	"github.com/google/uuid"
)

var (
	ErrTagNotFound = errors.New("tag not found")
	ErrTagExists   = errors.New("tag already exists")
)

type TagRepository interface {
	Create(ctx context.Context, tag *models.Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tag, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Tag, error)
	Update(ctx context.Context, tag *models.Tag) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TagService struct {
	repo TagRepository
}

func NewTagService(repo TagRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) Create(ctx context.Context, userID uuid.UUID, req models.CreateTagRequest) (*models.Tag, error) {
	color := req.Color
	if color == "" {
		color = "#6b7280"
	}

	tag := &models.Tag{
		UserID: userID,
		Name:   req.Name,
		Color:  color,
	}

	if err := s.repo.Create(ctx, tag); err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *TagService) List(ctx context.Context, userID uuid.UUID) ([]models.Tag, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *TagService) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, req models.UpdateTagRequest) (*models.Tag, error) {
	tag, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTagNotFound
	}
	if tag.UserID != userID {
		return nil, ErrTagNotFound
	}

	if req.Name != nil {
		tag.Name = *req.Name
	}
	if req.Color != nil {
		tag.Color = *req.Color
	}

	if err := s.repo.Update(ctx, tag); err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *TagService) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	tag, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrTagNotFound
	}
	if tag.UserID != userID {
		return ErrTagNotFound
	}

	return s.repo.Delete(ctx, id)
}
