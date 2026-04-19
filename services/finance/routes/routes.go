package routes

import (
	"github.com/DB-Vincent/personal-finance/services/finance/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func New(
	categories *handler.CategoryHandler,
	accounts *handler.AccountHandler,
	tags *handler.TagHandler,
	transactions *handler.TransactionHandler,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/categories", func(r chi.Router) {
		r.Get("/", categories.List)
		r.Post("/", categories.Create)
		r.Put("/{id}", categories.Update)
		r.Post("/{id}/archive", categories.ToggleArchive)
	})

	r.Route("/accounts", func(r chi.Router) {
		r.Get("/", accounts.List)
		r.Post("/", accounts.Create)
		r.Get("/net-worth", accounts.NetWorth)
		r.Get("/{id}", accounts.Get)
		r.Put("/{id}", accounts.Update)
		r.Post("/{id}/archive", accounts.ToggleArchive)
		r.Delete("/{id}", accounts.Delete)
	})

	r.Route("/tags", func(r chi.Router) {
		r.Get("/", tags.List)
		r.Post("/", tags.Create)
		r.Put("/{id}", tags.Update)
		r.Delete("/{id}", tags.Delete)
	})

	r.Route("/transactions", func(r chi.Router) {
		r.Get("/", transactions.List)
		r.Post("/", transactions.Create)
		r.Get("/{id}", transactions.Get)
		r.Put("/{id}", transactions.Update)
		r.Delete("/{id}", transactions.Delete)
	})

	return r
}
