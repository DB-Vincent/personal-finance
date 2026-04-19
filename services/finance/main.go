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
	"github.com/DB-Vincent/personal-finance/services/finance/config"
	"github.com/DB-Vincent/personal-finance/services/finance/handler"
	"github.com/DB-Vincent/personal-finance/services/finance/repository"
	"github.com/DB-Vincent/personal-finance/services/finance/routes"
	"github.com/DB-Vincent/personal-finance/services/finance/service"
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

	if err := database.RunMigrations(pool, migrations, "finance"); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	categoryRepo := repository.NewPostgresCategoryRepository(pool)
	accountRepo := repository.NewPostgresAccountRepository(pool)
	tagRepo := repository.NewPostgresTagRepository(pool)
	transactionRepo := repository.NewPostgresTransactionRepository(pool)
	transactionTagRepo := repository.NewPostgresTransactionTagRepository(pool)

	categorySvc := service.NewCategoryService(categoryRepo)
	accountSvc := service.NewAccountService(accountRepo)
	tagSvc := service.NewTagService(tagRepo)
	transactionSvc := service.NewTransactionService(transactionRepo, transactionTagRepo)

	categoryHandler := handler.NewCategoryHandler(categorySvc)
	accountHandler := handler.NewAccountHandler(accountSvc)
	tagHandler := handler.NewTagHandler(tagSvc)
	transactionHandler := handler.NewTransactionHandler(transactionSvc)

	router := routes.New(categoryHandler, accountHandler, tagHandler, transactionHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("finance service starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down finance service")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
