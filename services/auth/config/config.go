package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port                int    `env:"PORT" envDefault:"8081"`
	DatabaseURL         string `env:"DATABASE_URL,required"`
	JWTAccessSecret     string `env:"JWT_ACCESS_SECRET,required"`
	JWTRefreshSecret    string `env:"JWT_REFRESH_SECRET,required"`
	RegistrationEnabled bool   `env:"REGISTRATION_ENABLED" envDefault:"true"`
	AdminEmail          string `env:"ADMIN_EMAIL" envDefault:"admin@example.com"`
	AdminPassword       string `env:"ADMIN_PASSWORD" envDefault:"changeme"`
	LogLevel            string `env:"LOG_LEVEL" envDefault:"info"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
