package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/DB-Vincent/personal-finance/services/auth/models"
	"github.com/DB-Vincent/personal-finance/services/auth/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role)
		 VALUES ($1, $2, $3)
		 RETURNING id, display_name, currency_symbol, is_disabled, create_time, update_time`,
		user.Email, user.PasswordHash, user.Role,
	).Scan(&user.ID, &user.DisplayName, &user.CurrencySymbol, &user.IsDisabled, &user.CreateTime, &user.UpdateTime)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return service.ErrEmailExists
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, display_name, currency_symbol, role, is_disabled, create_time, update_time
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.CurrencySymbol, &user.Role, &user.IsDisabled, &user.CreateTime, &user.UpdateTime)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("query user by email: %w", err)
	}
	return user, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, display_name, currency_symbol, role, is_disabled, create_time, update_time
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.CurrencySymbol, &user.Role, &user.IsDisabled, &user.CreateTime, &user.UpdateTime)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("query user by id: %w", err)
	}
	return user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET display_name = $1, currency_symbol = $2, update_time = NOW()
		 WHERE id = $3`,
		user.DisplayName, user.CurrencySymbol, user.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}
