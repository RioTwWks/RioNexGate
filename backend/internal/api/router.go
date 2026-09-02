package api

import (
	"net/http"

	"rionexgate/internal/api/middleware"
	"rionexgate/internal/config"
	"rionexgate/internal/core"
	"rionexgate/internal/db"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(cfg *config.Config, database *db.DB, coreMgr core.Manager) http.Handler {
	h := NewHandler(database, coreMgr, cfg)
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.APIVersioning)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key", "X-Device-Token", "X-API-Version"},
		ExposedHeaders:   []string{"X-API-Version", "X-Config-Cached"},
		AllowCredentials: false,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", h.Health)
		r.Get("/docs", h.SwaggerUI)
		r.Get("/openapi.yaml", h.OpenAPISpec)
		r.Get("/subscription/{token}", h.GetSubscription)

		r.Route("/client", func(r chi.Router) {
			r.Post("/register", h.RegisterClient)
			r.Group(func(r chi.Router) {
				r.Use(middleware.DeviceTokenAuth(database))
				r.Use(middleware.ClientRequestLogger)
				r.Get("/config", h.GetClientConfig)
				r.Post("/stats", h.PostClientStats)
				r.Get("/commands", h.GetClientCommands)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.APIKeyAuth(cfg.Server.APIKey))
			r.Get("/protocols", h.ListProtocols)
			r.Get("/users", h.ListUsers)
			r.Post("/users", h.CreateUser)
			r.Get("/users/{id}", h.GetUser)
			r.Put("/users/{id}", h.UpdateUser)
			r.Delete("/users/{id}", h.DeleteUser)
			r.Get("/users/{id}/link", h.GetUserLink)
			r.Get("/users/{id}/qr", h.GetUserQR)
			r.Get("/stats/total", h.GetTotalStats)
			r.Get("/stats/user/{id}", h.GetUserStats)
			r.Get("/stats/client", h.GetClientMetrics)
			r.Post("/core/reload", h.ReloadCore)
			r.Get("/core/type", h.GetCoreType)
			r.Put("/core/type", h.SetCoreType)
		})
	})

	return r
}
