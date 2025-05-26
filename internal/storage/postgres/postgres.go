package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/storage"
	"github.com/theartofdevel/logging"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
)

type Storage struct {
	db  *sql.DB
	cfg *config.Config
}

func New(ctx context.Context, cfg *config.Config) (*Storage, error) {
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

	if _, err := db.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", cfg.Database.Schema)); err != nil {
		return nil, fmt.Errorf("%s: failed to set search_path: %w", op, err)
	}

	// Проверка соединения с таймаутом
	pingCtx, cancel := context.WithTimeout(ctx, cfg.Database.Timeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		logging.L(ctx).Error("postgres ping failed", sl.Err(err))
		return nil, fmt.Errorf("%s: ping failed: %w", op, err)
	}

	storageBD := &Storage{
		db:  db,
		cfg: cfg,
	}

	if err := storageBD.initSchema(ctx); err != nil {
		logging.L(ctx).Error("failed to init database schema", sl.Err(err))
		_ = db.Close()
		return nil, fmt.Errorf("%s: init schema failed: %w", op, err)
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
    		remove boolean,
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

func (s *Storage) GetVideosByCategoryAndType(ctx context.Context, contentType, category string) ([]response.VideoResponse, error) {
	const op = "storage.postgres.GetVideosByCategoryAndType"

	// Проверка существования типа контента
	err := s.chekType(ctx, contentType, op)
	if err != nil {
		return nil, err
	}

	// Проверка существования категории
	err = s.chekCategory(ctx, category, op)
	if err != nil {
		return nil, err
	}

	query := `
        SELECT v.id, v.url, v.name, v.description, c.name as category, v.img_url
        FROM videos v
        JOIN video_categories vc ON v.id = vc.video_id
        JOIN categories c ON vc.category_id = c.id
        JOIN category_content_types cct ON c.id = cct.category_id
        JOIN content_types ct ON cct.content_type_id = ct.id
        WHERE ct.id = $1 AND c.id = $2 AND v.deleted IS NOT TRUE
        ORDER BY v.created_at DESC
    `

	rows, err := s.db.QueryContext(ctx, query, contentType, category)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()

	var videos []response.VideoResponse
	for rows.Next() {
		var v response.VideoResponse
		if err := rows.Scan(&v.ID, &v.URL, &v.Name, &v.Description, &v.Category, &v.ImgURL); err != nil {
			return nil, fmt.Errorf("%s: scan failed: %w", op, err)
		}
		v.URL = s.constructFullMediaURL(v.URL)
		v.ImgURL = s.constructFullImgURL(v.ImgURL)
		videos = append(videos, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	// Проверка на отсутствие категорий для данного типа
	if len(videos) == 0 {
		return nil, fmt.Errorf("%s: %w: no videos found for content type '%s' and category '%s'",
			op, storage.ErrVideoNotFound, contentType, category)
	}

	return videos, nil
}

// CheckAccount возвращает Type ID для указанного username
func (s *Storage) CheckAccount(ctx context.Context, username string) (response.AccountResponse, error) {
	const op = "storage.postgres.CheckAccount"

	query := `
        SELECT a.content_type_id, ct.name
        FROM accounts a
        JOIN content_types ct ON a.content_type_id = ct.id 
        WHERE a.username = $1 AND a.deleted IS NOT TRUE
    `

	var account response.AccountResponse

	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&account.TypeID,
		&account.TypeName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.AccountResponse{}, fmt.Errorf("%s: %w: video with id '%s' not found",
				op, storage.ErrAccountNotFound, username)
		}
		return response.AccountResponse{}, fmt.Errorf("%s: query failed: %w", op, err)
	}

	return account, nil
}

// GetCategories возвращает все категории для указанного типа контента
func (s *Storage) GetCategories(ctx context.Context, contentType string) ([]response.CategoryResponse, error) {
	const op = "storage.postgres.GetCategories"

	// Проверка существования типа контента
	err := s.chekType(ctx, contentType, op)
	if err != nil {
		return nil, err
	}

	// Основной запрос для получения категорий
	query := `
        SELECT c.id, c.name, c.img_url
        FROM categories c
        JOIN category_content_types cct ON c.id = cct.category_id
        JOIN content_types ct ON cct.content_type_id = ct.id
        WHERE ct.id = $1 AND c.deleted IS NOT TRUE
        ORDER BY c.created_at DESC
    `

	rows, err := s.db.QueryContext(ctx, query, contentType)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()

	var categories []response.CategoryResponse

	for rows.Next() {
		var category response.CategoryResponse
		if err := rows.Scan(&category.ID, &category.Name, &category.ImgURL); err != nil {
			return nil, fmt.Errorf("%s: scan failed: %w", op, err)
		}
		category.ImgURL = s.constructFullImgURL(category.ImgURL)

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	// Проверка на отсутствие категорий для данного типа
	if len(categories) == 0 {
		return nil, fmt.Errorf("%s: %w: no categories found for content type '%s'",
			op, storage.ErrNoCategoriesFound, contentType)
	}

	return categories, nil
}

func (s *Storage) GetVideo(ctx context.Context, videoID string) (response.VideoResponse, error) {
	const op = "storage.postgres.GetVideo"

	query := `
        SELECT v.id, v.url, v.name, v.description, c.name as category, v.img_url
        FROM videos v
        JOIN video_categories vc ON v.id = vc.video_id
        JOIN categories c ON vc.category_id = c.id
        WHERE v.id = $1 AND v.deleted IS NOT TRUE
    `

	var video response.VideoResponse

	err := s.db.QueryRowContext(ctx, query, videoID).Scan(
		&video.ID,
		&video.URL,
		&video.Name,
		&video.Description,
		&video.Category,
		&video.ImgURL,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.VideoResponse{}, fmt.Errorf("%s: %w: video with id '%s' not found",
				op, storage.ErrVideoNotFound, videoID)
		}
		return response.VideoResponse{}, fmt.Errorf("%s: query failed: %w", op, err)
	}

	video.URL = s.constructFullMediaURL(video.URL)

	video.ImgURL = s.constructFullImgURL(video.ImgURL)

	return video, nil
}

func (s *Storage) chekType(ctx context.Context, contentType, op string) error {
	var contentTypeExists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM content_types WHERE id = $1 AND deleted IS NOT TRUE)`,
		contentType,
	).Scan(&contentTypeExists)

	if err != nil {
		return fmt.Errorf("%s: content type check failed: %w", op, err)
	}

	if !contentTypeExists {
		return fmt.Errorf("%s: %w: content type '%s' not found",
			op, storage.ErrContentTypeNotFound, contentType)
	}

	return nil
}

func (s *Storage) chekCategory(ctx context.Context, category, op string) error {
	var categoryNameExists bool

	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1 AND deleted IS NOT TRUE)`,
		category,
	).Scan(&categoryNameExists)

	if err != nil {
		return fmt.Errorf("%s: category name check failed: %w", op, err)
	}

	if !categoryNameExists {
		return fmt.Errorf("%s: %w: category name '%s' not found",
			op, storage.ErrNoCategoriesFound, category)
	}

	return nil
}

func (s *Storage) constructFullMediaURL(relativePath string) string {
	if relativePath == "" {
		return ""
	}

	baseURL := strings.TrimRight(s.cfg.Media.BaseURL, "/")
	videoPath := strings.TrimLeft(relativePath, "/")

	return fmt.Sprintf("%s/video/%s", baseURL, videoPath)
}

func (s *Storage) constructFullImgURL(relativePath string) string {
	if relativePath == "" {
		return ""
	}

	baseURL := strings.TrimRight(s.cfg.Media.BaseURL, "/")
	imgPath := strings.TrimLeft(relativePath, "/")

	return fmt.Sprintf("%s/img/%s", baseURL, imgPath)
}

func (s *Storage) Close() error {
	const op = "storage.postgres.Close"

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
