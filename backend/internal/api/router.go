package api

import (
	"net/http"

	"proxy-mgr/internal/api/middleware"
	"proxy-mgr/internal/config"
	"proxy-mgr/internal/core"
	"proxy-mgr/internal/db"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(cfg *config.Config, database *db.DB, coreMgr core.Manager) http.Handler {
	h := NewHandler(database, coreMgr, cfg)
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		AllowCredentials: false,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", h.Health)
		r.Group(func(r chi.Router) {
			r.Use(middleware.APIKeyAuth(cfg.Server.APIKey))
			r.Get("/users", h.ListUsers)
			r.Post("/users", h.CreateUser)
			r.Get("/users/{id}", h.GetUser)
			r.Put("/users/{id}", h.UpdateUser)
			r.Delete("/users/{id}", h.DeleteUser)
			r.Get("/users/{id}/link", h.GetUserLink)
			r.Get("/users/{id}/qr", h.GetUserQR)
			r.Get("/stats/total", h.GetTotalStats)
			r.Get("/stats/user/{id}", h.GetUserStats)
			r.Post("/core/reload", h.ReloadCore)
			r.Get("/core/type", h.GetCoreType)
			r.Put("/core/type", h.SetCoreType)
		})
	})

	return r
}
