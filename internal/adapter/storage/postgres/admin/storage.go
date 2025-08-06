package admin

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Storage {
	return &Storage{db: db}
}
