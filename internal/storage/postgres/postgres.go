package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/langowen/bodybalance-backend/internal/storage/postgres/admin"
	"github.com/langowen/bodybalance-backend/internal/storage/postgres/api"
	"github.com/theartofdevel/logging"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
)

type Storage struct {
	db    *sql.DB
	cfg   *config.Config
	Admin *admin.Storage
	Api   *api.Storage
}

func New(ctx context.Context, cfg *config.Config) (*Storage, error) {
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

	dbConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: parse config failed: %w", op, err)
	}

	dbConfig.ConnectTimeout = 5 * time.Second
	db := stdlib.OpenDB(*dbConfig)

	// Настройки пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Проверка соединения с таймаутом
	pingCtx, cancel := context.WithTimeout(ctx, cfg.Database.Timeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		logging.L(ctx).Error("postgres ping failed", sl.Err(err))
		return nil, fmt.Errorf("%s: ping failed: %w", op, err)
	}

	storageBD := &Storage{
		db:    db,
		cfg:   cfg,
		Admin: admin.New(db),
		Api:   api.New(db, cfg),
	}

	if err := storageBD.initSchema(ctx); err != nil {
		logging.L(ctx).Error("failed to init database schema", sl.Err(err))
		_ = db.Close()
		return nil, fmt.Errorf("%s: init schema failed: %w", op, err)
	}

	err = storageBD.InitData(ctx)
	if err != nil {
		logging.L(ctx).Error("failed create data to database", sl.Err(err))
	}

	logging.L(ctx).Info("PostgreSQL storage initialized successfully")
	return storageBD, nil
}

func (s *Storage) initSchema(ctx context.Context) error {
	const op = "storage.postgres.initSchema"

	queries := []string{
		`CREATE TABLE IF NOT EXISTS content_types (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
    		deleted boolean,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			img_url TEXT,
    		deleted BOOLEAN,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS category_content_types (
			category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
			content_type_id INTEGER REFERENCES content_types(id) ON DELETE CASCADE,
			PRIMARY KEY (category_id, content_type_id)
		)`,
		`CREATE TABLE IF NOT EXISTS videos (
			id SERIAL PRIMARY KEY,
			url TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			img_url TEXT,
    		deleted BOOLEAN,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS video_categories (
			video_id INTEGER REFERENCES videos(id) ON DELETE CASCADE,
			category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
			PRIMARY KEY (video_id, category_id)
		)`,
		`CREATE TABLE IF NOT EXISTS accounts (
			id SERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			content_type_id INTEGER REFERENCES content_types(id) ON DELETE SET NULL,
    		admin BOOLEAN,
    		password TEXT,
    		deleted BOOLEAN,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_video_categories_video_id ON video_categories(video_id)`,
		`CREATE INDEX IF NOT EXISTS idx_video_categories_category_id ON video_categories(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_category_content_types_category_id ON category_content_types(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_category_content_types_content_type_id ON category_content_types(content_type_id)`,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: begin transaction failed: %w", op, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("%s: exec query failed: %w", op, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit transaction failed: %w", op, err)
	}

	return nil
}

func (s *Storage) InitData(ctx context.Context) error {
	const op = "storage.postgres.initData"

	// Проверяем, есть ли уже тип 'admin'
	var typeID int

	err := s.db.QueryRowContext(ctx,
		"SELECT id FROM content_types WHERE name = 'admin'").Scan(&typeID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: failed to check admin type: %w", op, err)
	}

	// Если запись не найдена, создаём новую
	if errors.Is(err, sql.ErrNoRows) {
		err = s.db.QueryRowContext(ctx,
			"INSERT INTO content_types (name, deleted) VALUES ($1, $2) RETURNING id",
			"admin", false).Scan(&typeID)
		if err != nil {
			return fmt.Errorf("%s: create admin type failed: %w", op, err)
		}
	}

	// Проверяем, есть ли уже админ
	var exists bool
	err = s.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM accounts WHERE username = $1)",
		s.cfg.Docs.User,
	).Scan(&exists)

	if err != nil {
		return fmt.Errorf("%s: failed to check admin existence: %w", op, err)
	}

	// Если администратора нет — создаём
	if !exists {
		hashedPassword := hashPassword(s.cfg.Docs.Password)

		_, err = s.db.ExecContext(ctx,
			"INSERT INTO accounts (username, content_type_id, password, admin, deleted) VALUES ($1, $2, $3, $4, $5)",
			s.cfg.Docs.User, typeID, hashedPassword, true, false,
		)

		if err != nil {
			return fmt.Errorf("%s: create admin failed: %w", op, err)
		}
	}

	return nil
}

// Пример функции хеширования пароля
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func (s *Storage) Close() error {
	const op = "storage.postgres.Close"

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
