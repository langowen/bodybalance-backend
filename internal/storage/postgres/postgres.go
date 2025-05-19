package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/langowen/bodybalance-backend/internal/storage"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
)

type Storage struct {
	db     *sql.DB
	logger *logging.Logger
	cfg    *config.Config
}

func New(ctx context.Context, cfg *config.Config, logger *logging.Logger) (*Storage, error) {
	const op = "storage.postgres.New"

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
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
		logger.Error("postgres ping failed", sl.Err(err))
		return nil, fmt.Errorf("%s: ping failed: %w", op, err)
	}

	storageBD := &Storage{
		db:     db,
		logger: logger,
		cfg:    cfg,
	}

	if err := storageBD.initSchema(ctx); err != nil {
		logger.Error("failed to init database schema", sl.Err(err))
		_ = db.Close()
		return nil, fmt.Errorf("%s: init schema failed: %w", op, err)
	}

	logger.Info("PostgreSQL storage initialized successfully")
	return storageBD, nil
}

func (s *Storage) initSchema(ctx context.Context) error {
	const op = "storage.postgres.initSchema"

	queries := []string{
		`CREATE TABLE IF NOT EXISTS content_types (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
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

func (s *Storage) Close() error {
	const op = "storage.postgres.Close"

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetVideosByCategoryAndType(ctx context.Context, contentType, categoryName string) ([]storage.Video, error) {
	const op = "storage.postgres.GetVideosByCategoryAndType"

	query := `
        SELECT v.id, v.url, v.name, v.description, c.name as category
        FROM videos v
        JOIN video_categories vc ON v.id = vc.video_id
        JOIN categories c ON vc.category_id = c.id
        JOIN category_content_types cct ON c.id = cct.category_id
        JOIN content_types ct ON cct.content_type_id = ct.id
        WHERE LOWER(ct.name) = LOWER($1) AND LOWER(c.name) = LOWER($2)
        ORDER BY v.created_at DESC
    `

	rows, err := s.db.QueryContext(ctx, query, contentType, categoryName)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()

	var videos []storage.Video
	for rows.Next() {
		var v storage.Video
		if err := rows.Scan(&v.ID, &v.URL, &v.Name, &v.Description, &v.Category); err != nil {
			return nil, fmt.Errorf("%s: scan failed: %w", op, err)
		}
		v.URL = s.constructFullMediaURL(v.URL)
		videos = append(videos, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return videos, nil
}

func (s *Storage) CheckAccountType(ctx context.Context, username, contentType string) (bool, error) {
	const op = "storage.postgres.CheckAccountType"

	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM accounts a
			JOIN content_types ct ON a.content_type_id = ct.id
			WHERE LOWER(a.username) = LOWER($1) AND LOWER(ct.name) = LOWER($2)
		)
	`

	var exists bool
	if err := s.db.QueryRowContext(ctx, query, username, contentType).Scan(&exists); err != nil {
		return false, fmt.Errorf("%s: query failed: %w", op, err)
	}

	return exists, nil
}

// GetCategoriesWithVideos возвращает все категории с видео для указанного типа
func (s *Storage) GetCategoriesWithVideos(ctx context.Context, contentType string) ([]storage.CategoryWithVideos, error) {
	const op = "storage.postgres.GetCategoriesWithVideos"

	query := `
		WITH cat AS (
			SELECT c.id, c.name 
			FROM categories c
			JOIN category_content_types cct ON c.id = cct.category_id
			JOIN content_types ct ON cct.content_type_id = ct.id
			WHERE LOWER(ct.name) = LOWER($1)
		)
		SELECT 
			cat.id as category_id,
			cat.name as category_name,
			v.id as video_id,
			v.url,
			v.name as video_name,
			v.description
		FROM cat
		LEFT JOIN video_categories vc ON cat.id = vc.category_id
		LEFT JOIN videos v ON vc.video_id = v.id
		ORDER BY cat.name, v.created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, contentType)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()

	var results []storage.CategoryWithVideos
	var currentCategory *storage.CategoryWithVideos

	for rows.Next() {
		var (
			catID     int
			catName   string
			videoID   sql.NullInt64
			videoURL  sql.NullString
			videoName sql.NullString
			videoDesc sql.NullString
		)

		if err := rows.Scan(&catID, &catName, &videoID, &videoURL, &videoName, &videoDesc); err != nil {
			return nil, fmt.Errorf("%s: scan failed: %w", op, err)
		}

		// Если это новая категория
		if currentCategory == nil || currentCategory.ID != catID {
			if currentCategory != nil {
				results = append(results, *currentCategory)
			}
			currentCategory = &storage.CategoryWithVideos{
				ID:     catID,
				Name:   catName,
				Videos: []storage.Video{},
			}
		}

		// Добавляем видео, если оно есть
		if videoID.Valid {
			currentCategory.Videos = append(currentCategory.Videos, storage.Video{
				ID:          float64(videoID.Int64),
				URL:         s.constructFullMediaURL(videoURL.String),
				Name:        videoName.String,
				Description: videoDesc.String,
				Category:    catName,
			})
		}
	}

	// Добавляем последнюю категорию
	if currentCategory != nil {
		results = append(results, *currentCategory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return results, nil
}

func (s *Storage) constructFullMediaURL(relativePath string) string {
	if relativePath == "" {
		return ""
	}

	baseURL := strings.TrimRight(s.cfg.Media.BaseURL, "/")
	videoPath := strings.TrimLeft(relativePath, "/")

	return fmt.Sprintf("%s/video/%s", baseURL, videoPath)
}
