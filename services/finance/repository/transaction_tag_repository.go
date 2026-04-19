package repository

import (
	"context"
	"fmt"

	"github.com/DB-Vincent/personal-finance/services/finance/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTransactionTagRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTransactionTagRepository(pool *pgxpool.Pool) *PostgresTransactionTagRepository {
	return &PostgresTransactionTagRepository{pool: pool}
}

func (r *PostgresTransactionTagRepository) SetTags(ctx context.Context, transactionID uuid.UUID, tagIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM transaction_tags WHERE transaction_id = $1`, transactionID)
	if err != nil {
		return fmt.Errorf("delete existing tags: %w", err)
	}

	for _, tagID := range tagIDs {
		_, err = tx.Exec(ctx,
			`INSERT INTO transaction_tags (transaction_id, tag_id) VALUES ($1, $2)`,
			transactionID, tagID,
		)
		if err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresTransactionTagRepository) GetTags(ctx context.Context, transactionID uuid.UUID) ([]models.Tag, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT t.id, t.user_id, t.name, t.color, t.create_time
		 FROM tags t
		 JOIN transaction_tags tt ON tt.tag_id = t.id
		 WHERE tt.transaction_id = $1
		 ORDER BY t.name`,
		transactionID,
	)
	if err != nil {
		return nil, fmt.Errorf("get transaction tags: %w", err)
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
