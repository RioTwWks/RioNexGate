package api

import (
	"sync"

	"rionexgate/internal/config"
	"rionexgate/internal/core"
	"rionexgate/internal/db"
)

type Handler struct {
	db       *db.DB
	core     core.Manager
	cfg      *config.Config
	commands *CommandManager
	metrics  ClientMetrics

	// publicBaseWarnOnce limits the "public_base_url not set" warning to one log line.
	publicBaseWarnOnce sync.Once
}

func NewHandler(database *db.DB, coreMgr core.Manager, cfg *config.Config) *Handler {
	return &Handler{
		db:       database,
		core:     coreMgr,
		cfg:      cfg,
		commands: NewCommandManager(),
	}
}
