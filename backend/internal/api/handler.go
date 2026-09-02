package api

import (
	"rionexgate/internal/config"
	"rionexgate/internal/core"
	"rionexgate/internal/db"
)

type Handler struct {
	db   *db.DB
	core core.Manager
	cfg  *config.Config
}

func NewHandler(database *db.DB, coreMgr core.Manager, cfg *config.Config) *Handler {
	return &Handler{db: database, core: coreMgr, cfg: cfg}
}
