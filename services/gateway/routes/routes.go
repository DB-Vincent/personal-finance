package routes

import (
	"net/http"

	"github.com/DB-Vincent/personal-finance/services/gateway/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Config struct {
	AuthProxy       http.Handler
	FinanceProxy    http.Handler
	JWTSecret       []byte
	CORSOptions     cors.Options
	RateLimitPerSec int
}

func New(cfg Config) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(cors.Handler(cfg.CORSOptions))
	r.Use(middleware.RateLimit(cfg.RateLimitPerSec))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	authStripped := http.StripPrefix("/api/v1", cfg.AuthProxy)
	financeStripped := http.StripPrefix("/api/v1", cfg.FinanceProxy)

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/auth/register", authStripped.ServeHTTP)
			r.Post("/auth/login", authStripped.ServeHTTP)
			r.Post("/auth/refresh", authStripped.ServeHTTP)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWTSecret))
			r.Handle("/auth/*", authStripped)
			r.Handle("/users/*", authStripped)
			r.Handle("/finance/*", financeStripped)
		})
	})

	return r
}
