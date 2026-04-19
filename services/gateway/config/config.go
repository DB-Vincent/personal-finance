package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port               int    `env:"PORT" envDefault:"8080"`
	AuthServiceURL     string `env:"AUTH_SERVICE_URL,required"`
	FinanceServiceURL  string `env:"FINANCE_SERVICE_URL" envDefault:"http://localhost:8082"`
	JWTAccessSecret    string `env:"JWT_ACCESS_SECRET,required"`
	CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:5173"`
	RateLimit          int    `env:"RATE_LIMIT" envDefault:"100"`
	LogLevel           string `env:"LOG_LEVEL" envDefault:"info"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
