package api

import (
	"proxy-mgr/internal/config"
	"proxy-mgr/internal/core"
	"proxy-mgr/internal/db"
)

type Handler struct {
	db   *db.DB
	core core.Manager
	cfg  *config.Config
}

func NewHandler(database *db.DB, coreMgr core.Manager, cfg *config.Config) *Handler {
	return &Handler{db: database, core: coreMgr, cfg: cfg}
}
