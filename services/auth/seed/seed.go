package seed

import (
	"context"
	"log/slog"

	"github.com/DB-Vincent/personal-finance/services/auth/models"
	"github.com/DB-Vincent/personal-finance/services/auth/service"
	"golang.org/x/crypto/bcrypt"
)

func AdminUser(ctx context.Context, repo service.UserRepository, email, password string) error {
	_, err := repo.GetByEmail(ctx, email)
	if err == nil {
		slog.Info("admin user already exists", "email", email)
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         "admin",
	}

	if err := repo.Create(ctx, user); err != nil {
		return err
	}

	slog.Info("admin user created", "email", email)
	return nil
}
