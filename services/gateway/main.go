package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DB-Vincent/personal-finance/pkg/logger"
	"github.com/DB-Vincent/personal-finance/services/gateway/config"
	"github.com/DB-Vincent/personal-finance/services/gateway/middleware"
	"github.com/DB-Vincent/personal-finance/services/gateway/proxy"
	"github.com/DB-Vincent/personal-finance/services/gateway/routes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(cfg.LogLevel)

	authProxy, err := proxy.NewServiceProxy(cfg.AuthServiceURL)
	if err != nil {
		slog.Error("failed to create auth proxy", "error", err)
		os.Exit(1)
	}

	financeProxy, err := proxy.NewServiceProxy(cfg.FinanceServiceURL)
	if err != nil {
		slog.Error("failed to create finance proxy", "error", err)
		os.Exit(1)
	}

	router := routes.New(routes.Config{
		AuthProxy:       authProxy,
		FinanceProxy:    financeProxy,
		JWTSecret:       []byte(cfg.JWTAccessSecret),
		CORSOptions:     middleware.CORS(cfg.CORSAllowedOrigins),
		RateLimitPerSec: cfg.RateLimit,
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("gateway starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down gateway")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
