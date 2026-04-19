package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/DB-Vincent/personal-finance/services/finance/models"
	"github.com/DB-Vincent/personal-finance/services/finance/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCategoryRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresCategoryRepository(pool *pgxpool.Pool) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{pool: pool}
}

func (r *PostgresCategoryRepository) Create(ctx context.Context, cat *models.Category) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO categories (user_id, group_name, name, is_income)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, create_time`,
		cat.UserID, cat.GroupName, cat.Name, cat.IsIncome,
	).Scan(&cat.ID, &cat.CreateTime)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return service.ErrCategoryExists
		}
		return fmt.Errorf("insert category: %w", err)
	}

	return nil
}

func (r *PostgresCategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	cat := &models.Category{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, group_name, name, is_income, is_archived, create_time
		 FROM categories WHERE id = $1`,
		id,
	).Scan(&cat.ID, &cat.UserID, &cat.GroupName, &cat.Name, &cat.IsIncome, &cat.IsArchived, &cat.CreateTime)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category: %w", err)
	}

	return cat, nil
}

func (r *PostgresCategoryRepository) ListByUser(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]models.Category, error) {
	query := `SELECT id, user_id, group_name, name, is_income, is_archived, create_time
			  FROM categories WHERE user_id = $1`
	if !includeArchived {
		query += ` AND is_archived = false`
	}
	query += ` ORDER BY group_name, name`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var cats []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.UserID, &cat.GroupName, &cat.Name, &cat.IsIncome, &cat.IsArchived, &cat.CreateTime); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		cats = append(cats, cat)
	}

	return cats, nil
}

func (r *PostgresCategoryRepository) Update(ctx context.Context, cat *models.Category) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE categories SET group_name = $1, name = $2 WHERE id = $3`,
		cat.GroupName, cat.Name, cat.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return service.ErrCategoryExists
		}
		return fmt.Errorf("update category: %w", err)
	}
	if result.RowsAffected() == 0 {
		return service.ErrCategoryNotFound
	}
	return nil
}

func (r *PostgresCategoryRepository) SetArchived(ctx context.Context, id uuid.UUID, archived bool) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE categories SET is_archived = $1 WHERE id = $2`,
		archived, id,
	)
	if err != nil {
		return fmt.Errorf("set archived: %w", err)
	}
	if result.RowsAffected() == 0 {
		return service.ErrCategoryNotFound
	}
	return nil
}

func (r *PostgresCategoryRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM categories WHERE user_id = $1`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count categories: %w", err)
	}
	return count, nil
}
