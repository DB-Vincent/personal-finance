package routes

import (
	"github.com/DB-Vincent/personal-finance/services/auth/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func New(auth *handler.AuthHandler, user *handler.UserHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", auth.Register)
		r.Post("/login", auth.Login)
		r.Post("/refresh", auth.Refresh)
	})

	r.Route("/users", func(r chi.Router) {
		r.Get("/me", user.GetProfile)
		r.Put("/me", user.UpdateProfile)
	})

	return r
}
