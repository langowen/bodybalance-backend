package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/internal/adapter/storage/postgres/admin"
	"github.com/langowen/bodybalance-backend/internal/adapter/storage/postgres/api"
	"github.com/theartofdevel/logging"
	"time"

	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
)

type Storage struct {
	db    *pgxpool.Pool
	cfg   *config.Config
	Admin *admin.Storage
	Api   *api.Storage
}

func NewStorage(db *pgxpool.Pool, cfg *config.Config) *Storage {
	return &Storage{
		db:    db,
		cfg:   cfg,
		Admin: admin.New(db),
		Api:   api.New(db, cfg),
	}
}

func InitStorage(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	const op = "storage.postgres.New"

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
		cfg.Database.Schema,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: parse config failed: %w", op, err)
	}
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 10 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	pingCtx, cancel := context.WithTimeout(ctx, cfg.Database.Timeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(pingCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = pool.Ping(ctx); err != nil {
		logging.L(ctx).Error("postgres ping failed", sl.Err(err))
		return nil, fmt.Errorf("%s: ping failed: %w", op, err)
	}

	logging.L(ctx).Info("PostgreSQL storage initialized successfully")
	return pool, nil
}
