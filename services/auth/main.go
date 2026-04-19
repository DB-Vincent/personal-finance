package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DB-Vincent/personal-finance/pkg/database"
	"github.com/DB-Vincent/personal-finance/pkg/logger"
	"github.com/DB-Vincent/personal-finance/services/auth/config"
	"github.com/DB-Vincent/personal-finance/services/auth/handler"
	"github.com/DB-Vincent/personal-finance/services/auth/repository"
	"github.com/DB-Vincent/personal-finance/services/auth/routes"
	"github.com/DB-Vincent/personal-finance/services/auth/seed"
	"github.com/DB-Vincent/personal-finance/services/auth/service"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(cfg.LogLevel)

	ctx := context.Background()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := database.RunMigrations(pool, migrations, "auth"); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	repo := repository.NewPostgresUserRepository(pool)
	tokenSvc := service.NewTokenService(cfg.JWTAccessSecret, cfg.JWTRefreshSecret)
	authSvc := service.NewAuthService(repo, tokenSvc, cfg.RegistrationEnabled)

	if err := seed.AdminUser(ctx, repo, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		slog.Error("failed to seed admin user", "error", err)
		os.Exit(1)
	}

	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(authSvc)
	router := routes.New(authHandler, userHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("auth service starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down auth service")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
