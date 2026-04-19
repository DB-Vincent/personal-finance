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

type PostgresTagRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTagRepository(pool *pgxpool.Pool) *PostgresTagRepository {
	return &PostgresTagRepository{pool: pool}
}

func (r *PostgresTagRepository) Create(ctx context.Context, tag *models.Tag) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tags (user_id, name, color)
		 VALUES ($1, $2, $3)
		 RETURNING id, create_time`,
		tag.UserID, tag.Name, tag.Color,
	).Scan(&tag.ID, &tag.CreateTime)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return service.ErrTagExists
		}
		return fmt.Errorf("insert tag: %w", err)
	}

	return nil
}

func (r *PostgresTagRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	tag := &models.Tag{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, color, create_time FROM tags WHERE id = $1`,
		id,
	).Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreateTime)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrTagNotFound
		}
		return nil, fmt.Errorf("get tag: %w", err)
	}

	return tag, nil
}

func (r *PostgresTagRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Tag, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, color, create_time FROM tags WHERE user_id = $1 ORDER BY name`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreateTime); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *PostgresTagRepository) Update(ctx context.Context, tag *models.Tag) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE tags SET name = $1, color = $2 WHERE id = $3`,
		tag.Name, tag.Color, tag.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return service.ErrTagExists
		}
		return fmt.Errorf("update tag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return service.ErrTagNotFound
	}
	return nil
}

func (r *PostgresTagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM tags WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return service.ErrTagNotFound
	}
	return nil
}
