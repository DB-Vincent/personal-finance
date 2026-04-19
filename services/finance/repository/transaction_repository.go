package repository

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/DB-Vincent/personal-finance/services/finance/models"
	"github.com/DB-Vincent/personal-finance/services/finance/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTransactionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTransactionRepository(pool *pgxpool.Pool) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{pool: pool}
}

func (r *PostgresTransactionRepository) Create(ctx context.Context, tx *models.Transaction) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO transactions (user_id, account_id, type, amount, category_id, transfer_account_id, date, notes, recurring_rule_id, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id, create_time, update_time`,
		tx.UserID, tx.AccountID, tx.Type, tx.Amount, tx.CategoryID,
		tx.TransferAccountID, tx.Date, tx.Notes, tx.RecurringRuleID,
		tx.CreatedBy, tx.UpdatedBy,
	).Scan(&tx.ID, &tx.CreateTime, &tx.UpdateTime)

	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}

	return nil
}

func (r *PostgresTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	tx := &models.Transaction{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, account_id, type, amount, category_id, transfer_account_id,
				date, notes, recurring_rule_id, created_by, create_time, updated_by, update_time
		 FROM transactions WHERE id = $1`,
		id,
	).Scan(
		&tx.ID, &tx.UserID, &tx.AccountID, &tx.Type, &tx.Amount, &tx.CategoryID,
		&tx.TransferAccountID, &tx.Date, &tx.Notes, &tx.RecurringRuleID,
		&tx.CreatedBy, &tx.CreateTime, &tx.UpdatedBy, &tx.UpdateTime,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("get transaction: %w", err)
	}

	return tx, nil
}

func (r *PostgresTransactionRepository) List(ctx context.Context, userID uuid.UUID, filter models.TransactionFilter) ([]models.Transaction, string, int64, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("t.user_id = $%d", argIdx))
	args = append(args, userID)
	argIdx++

	if filter.AccountID != nil {
		conditions = append(conditions, fmt.Sprintf("(t.account_id = $%d OR t.transfer_account_id = $%d)", argIdx, argIdx))
		args = append(args, *filter.AccountID)
		argIdx++
	}
	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("t.category_id = $%d", argIdx))
		args = append(args, *filter.CategoryID)
		argIdx++
	}
	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("t.type = $%d", argIdx))
		args = append(args, *filter.Type)
		argIdx++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("t.date >= $%d", argIdx))
		args = append(args, *filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("t.date <= $%d", argIdx))
		args = append(args, *filter.DateTo)
		argIdx++
	}
	if filter.AmountMin != nil {
		conditions = append(conditions, fmt.Sprintf("t.amount >= $%d", argIdx))
		args = append(args, *filter.AmountMin)
		argIdx++
	}
	if filter.AmountMax != nil {
		conditions = append(conditions, fmt.Sprintf("t.amount <= $%d", argIdx))
		args = append(args, *filter.AmountMax)
		argIdx++
	}
	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("t.notes ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}
	if filter.TagID != nil {
		conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM transaction_tags tt WHERE tt.transaction_id = t.id AND tt.tag_id = $%d)", argIdx))
		args = append(args, *filter.TagID)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count total
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM transactions t WHERE %s", where)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, "", 0, fmt.Errorf("count transactions: %w", err)
	}

	// Cursor pagination
	if filter.PageToken != "" {
		cursorID, err := decodeCursor(filter.PageToken)
		if err == nil {
			conditions = append(conditions, fmt.Sprintf("(t.date, t.id) < (SELECT date, id FROM transactions WHERE id = $%d)", argIdx))
			args = append(args, cursorID)
			argIdx++
			where = strings.Join(conditions, " AND ")
		}
	}

	pageSize := filter.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50
	}

	query := fmt.Sprintf(
		`SELECT t.id, t.user_id, t.account_id, t.type, t.amount, t.category_id,
				t.transfer_account_id, t.date, t.notes, t.recurring_rule_id,
				t.created_by, t.create_time, t.updated_by, t.update_time
		 FROM transactions t
		 WHERE %s
		 ORDER BY t.date DESC, t.id DESC
		 LIMIT $%d`,
		where, argIdx,
	)
	args = append(args, pageSize+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", 0, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.AccountID, &tx.Type, &tx.Amount, &tx.CategoryID,
			&tx.TransferAccountID, &tx.Date, &tx.Notes, &tx.RecurringRuleID,
			&tx.CreatedBy, &tx.CreateTime, &tx.UpdatedBy, &tx.UpdateTime,
		); err != nil {
			return nil, "", 0, fmt.Errorf("scan transaction: %w", err)
		}
		txs = append(txs, tx)
	}

	var nextToken string
	if len(txs) > pageSize {
		txs = txs[:pageSize]
		nextToken = encodeCursor(txs[len(txs)-1].ID)
	}

	return txs, nextToken, total, nil
}

func (r *PostgresTransactionRepository) Update(ctx context.Context, tx *models.Transaction) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE transactions SET account_id = $1, type = $2, amount = $3, category_id = $4,
		 transfer_account_id = $5, date = $6, notes = $7, updated_by = $8, update_time = NOW()
		 WHERE id = $9`,
		tx.AccountID, tx.Type, tx.Amount, tx.CategoryID,
		tx.TransferAccountID, tx.Date, tx.Notes, tx.UpdatedBy, tx.ID,
	)
	if err != nil {
		return fmt.Errorf("update transaction: %w", err)
	}
	if result.RowsAffected() == 0 {
		return service.ErrTransactionNotFound
	}
	return nil
}

func (r *PostgresTransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	if result.RowsAffected() == 0 {
		return service.ErrTransactionNotFound
	}
	return nil
}

func encodeCursor(id uuid.UUID) string {
	return base64.URLEncoding.EncodeToString([]byte(id.String()))
}

func decodeCursor(token string) (uuid.UUID, error) {
	b, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(string(b))
}
