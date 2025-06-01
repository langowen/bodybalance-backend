package app

import (
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/storage/postgres"
	"github.com/theartofdevel/logging"
)

type App struct {
	Cfg     *config.Config
	Logger  *logging.Logger
	Storage *postgres.Storage
}

func New(cfg *config.Config, logger *logging.Logger, pgStorage *postgres.Storage) *App {
	return &App{
		Cfg:     cfg,
		Logger:  logger,
		Storage: pgStorage,
	}
}
