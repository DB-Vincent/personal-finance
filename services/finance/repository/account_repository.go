package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/DB-Vincent/personal-finance/services/finance/models"
	"github.com/DB-Vincent/personal-finance/services/finance/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type PostgresAccountRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAccountRepository(pool *pgxpool.Pool) *PostgresAccountRepository {
	return &PostgresAccountRepository{pool: pool}
}

func (r *PostgresAccountRepository) Create(ctx context.Context, acct *models.Account) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO accounts (user_id, name, type, starting_balance, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, create_time, update_time`,
		acct.UserID, acct.Name, acct.Type, acct.StartingBalance, acct.CreatedBy, acct.UpdatedBy,
	).Scan(&acct.ID, &acct.CreateTime, &acct.UpdateTime)

	if err != nil {
		return fmt.Errorf("insert account: %w", err)
	}

	return nil
}

const balanceQuery = `
	SELECT a.id, a.user_id, a.name, a.type, a.starting_balance, a.is_archived,
		   a.created_by, a.create_time, a.updated_by, a.update_time,
		   a.starting_balance + COALESCE(SUM(
			   CASE
				   WHEN t.type = 'income' AND t.account_id = a.id THEN t.amount
				   WHEN t.type = 'transfer' AND t.transfer_account_id = a.id THEN t.amount
				   WHEN t.type = 'expense' AND t.account_id = a.id THEN -t.amount
				   WHEN t.type = 'transfer' AND t.account_id = a.id THEN -t.amount
				   ELSE 0
			   END
		   ), 0) AS balance
	FROM accounts a
	LEFT JOIN transactions t ON (t.account_id = a.id OR t.transfer_account_id = a.id)
`

func scanAccount(row pgx.Row) (*models.Account, error) {
	acct := &models.Account{}
	err := row.Scan(
		&acct.ID, &acct.UserID, &acct.Name, &acct.Type, &acct.StartingBalance,
		&acct.IsArchived, &acct.CreatedBy, &acct.CreateTime, &acct.UpdatedBy,
		&acct.UpdateTime, &acct.Balance,
	)
	return acct, err
}

func (r *PostgresAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	query := balanceQuery + ` WHERE a.id = $1 GROUP BY a.id`
	acct, err := scanAccount(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrAccountNotFound
		}
		return nil, fmt.Errorf("get account: %w", err)
	}
	return acct, nil
}

func (r *PostgresAccountRepository) ListByUser(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]models.Account, error) {
	query := balanceQuery + ` WHERE a.user_id = $1`
	if !includeArchived {
		query += ` AND a.is_archived = false`
	}
	query += ` GROUP BY a.id ORDER BY a.name`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		acct := models.Account{}
		if err := rows.Scan(
			&acct.ID, &acct.UserID, &acct.Name, &acct.Type, &acct.StartingBalance,
			&acct.IsArchived, &acct.CreatedBy, &acct.CreateTime, &acct.UpdatedBy,
			&acct.UpdateTime, &acct.Balance,
		); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, acct)
	}

	return accounts, nil
}

func (r *PostgresAccountRepository) Update(ctx context.Context, acct *models.Account) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE accounts SET name = $1, type = $2, updated_by = $3, update_time = NOW()
		 WHERE id = $4`,
		acct.Name, acct.Type, acct.UpdatedBy, acct.ID,
	)
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	if result.RowsAffected() == 0 {
		return service.ErrAccountNotFound
	}
	return nil
}

func (r *PostgresAccountRepository) SetArchived(ctx context.Context, id uuid.UUID, archived bool) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE accounts SET is_archived = $1, update_time = NOW() WHERE id = $2`,
		archived, id,
	)
	if err != nil {
		return fmt.Errorf("set archived: %w", err)
	}
	if result.RowsAffected() == 0 {
		return service.ErrAccountNotFound
	}
	return nil
}

func (r *PostgresAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM accounts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	if result.RowsAffected() == 0 {
		return service.ErrAccountNotFound
	}
	return nil
}

func (r *PostgresAccountRepository) HasTransactions(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM transactions WHERE account_id = $1 OR transfer_account_id = $1)`,
		id,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check transactions: %w", err)
	}
	return exists, nil
}

func (r *PostgresAccountRepository) NetWorth(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	var total decimal.Decimal
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(balance), 0) FROM (
			SELECT a.starting_balance + COALESCE(SUM(
				CASE
					WHEN t.type = 'income' AND t.account_id = a.id THEN t.amount
					WHEN t.type = 'transfer' AND t.transfer_account_id = a.id THEN t.amount
					WHEN t.type = 'expense' AND t.account_id = a.id THEN -t.amount
					WHEN t.type = 'transfer' AND t.account_id = a.id THEN -t.amount
					ELSE 0
				END
			), 0) AS balance
			FROM accounts a
			LEFT JOIN transactions t ON (t.account_id = a.id OR t.transfer_account_id = a.id)
			WHERE a.user_id = $1 AND a.is_archived = false
			GROUP BY a.id
		) sub`,
		userID,
	).Scan(&total)
	if err != nil {
		return decimal.Zero, fmt.Errorf("net worth: %w", err)
	}
	return total, nil
}
