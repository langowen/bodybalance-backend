package api

import (
	"database/sql"
	"github.com/langowen/bodybalance-backend/internal/config"
)

type Storage struct {
	db  *sql.DB
	cfg *config.Config
}

func New(db *sql.DB, cfg *config.Config) *Storage {
	return &Storage{
		db:  db,
		cfg: cfg,
	}
}
