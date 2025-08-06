package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/internal/adapter/storage"
	"github.com/langowen/bodybalance-backend/internal/entities/api"
	"strings"
)

type Storage struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func New(db *pgxpool.Pool, cfg *config.Config) *Storage {
	return &Storage{
		db:  db,
		cfg: cfg,
	}
}

// CheckAccount возвращает Type ID для указанного username
func (s *Storage) CheckAccount(ctx context.Context, account *api.Account) (*api.Account, error) {
	const op = "storage.postgres.CheckAccount"

	query := `
        SELECT a.content_type_id, ct.name
        FROM accounts a
        JOIN content_types ct ON a.content_type_id = ct.id 
        WHERE a.username = $1 AND a.deleted IS NOT TRUE
    `

	row := s.db.QueryRow(ctx, query, account.Username)
	err := row.Scan(&account.ContentType.ID, &account.ContentType.Name)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w: video with id '%s' not found",
				op, storage.ErrAccountNotFound, account.Username)
		}
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}

	return account, nil
}

// GetCategories возвращает все категории для указанного типа контента
func (s *Storage) GetCategories(ctx context.Context, TypeID int64) ([]api.Category, error) {
	const op = "storage.postgres.GetCategories"

	err := s.chekType(ctx, TypeID, op)
	if err != nil {
		return nil, err
	}

	query := `
        SELECT c.id, c.name, c.img_url
        FROM categories c
        JOIN category_content_types cct ON c.id = cct.category_id
        JOIN content_types ct ON cct.content_type_id = ct.id
        WHERE ct.id = $1 AND c.deleted IS NOT TRUE
        ORDER BY c.created_at DESC
    `

	rows, err := s.db.Query(ctx, query, TypeID)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()

	var categories []api.Category

	for rows.Next() {
		var category api.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.ImgURL); err != nil {
			return nil, fmt.Errorf("%s: scan failed: %w", op, err)
		}
		category.ImgURL = s.constructFullImgURL(category.ImgURL)

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	if len(categories) == 0 {
		return nil, fmt.Errorf("%s: %w: no categories found for content type '%d'",
			op, storage.ErrNoCategoriesFound, TypeID)
	}

	return categories, nil
}

func (s *Storage) GetVideo(ctx context.Context, videoID int64) (*api.Video, error) {
	const op = "storage.postgres.GetVideo"

	query := `
        SELECT v.id, v.url, v.name, v.description, c.name as category, v.img_url
        FROM videos v
        JOIN video_categories vc ON v.id = vc.video_id
        JOIN categories c ON vc.category_id = c.id
        WHERE v.id = $1 AND v.deleted IS NOT TRUE
    `

	var video api.Video

	row := s.db.QueryRow(ctx, query, videoID)
	err := row.Scan(
		&video.ID,
		&video.URL,
		&video.Name,
		&video.Description,
		&video.Category.Name,
		&video.ImgURL,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w: video with id '%d' not found",
				op, storage.ErrVideoNotFound, videoID)
		}
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}

	video.URL = s.constructFullMediaURL(video.URL)

	video.ImgURL = s.constructFullImgURL(video.ImgURL)

	return &video, nil
}

func (s *Storage) GetVideosByCategoryAndType(ctx context.Context, TypeID, CatID int64) ([]api.Video, error) {
	const op = "storage.postgres.GetVideosByCategoryAndType"

	err := s.chekType(ctx, TypeID, op)
	if err != nil {
		return nil, err
	}

	err = s.chekCategory(ctx, CatID, op)
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

	rows, err := s.db.Query(ctx, query, TypeID, CatID)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()

	var videos []api.Video
	for rows.Next() {
		var v api.Video
		if err = rows.Scan(&v.ID, &v.URL, &v.Name, &v.Description, &v.Category.Name, &v.ImgURL); err != nil {
			return nil, fmt.Errorf("%s: scan failed: %w", op, err)
		}
		v.URL = s.constructFullMediaURL(v.URL)
		v.ImgURL = s.constructFullImgURL(v.ImgURL)
		videos = append(videos, v)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	if len(videos) == 0 {
		return nil, fmt.Errorf("%s: %w: no videos found for content type '%d' and category '%d'",
			op, storage.ErrVideoNotFound, TypeID, CatID)
	}

	return videos, nil
}

func (s *Storage) Feedback(ctx context.Context, feedback *api.Feedback) error {
	const op = "storage.postgres.Feedback"

	query := `
		INSERT INTO feedback (username, email, telegram, message)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int
	err := s.db.QueryRow(ctx, query,
		feedback.Name,
		feedback.Email,
		feedback.Telegram,
		feedback.Message,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) chekType(ctx context.Context, TypeID int64, op string) error {
	var contentTypeExists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM content_types WHERE id = $1 AND deleted IS NOT TRUE)`,
		TypeID,
	).Scan(&contentTypeExists)

	if err != nil {
		return fmt.Errorf("%s: content type check failed: %w", op, err)
	}

	if !contentTypeExists {
		return fmt.Errorf("%s: %w: content type '%d' not found",
			op, storage.ErrContentTypeNotFound, TypeID)
	}

	return nil
}

func (s *Storage) chekCategory(ctx context.Context, CatID int64, op string) error {
	var categoryNameExists bool

	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1 AND deleted IS NOT TRUE)`,
		CatID,
	).Scan(&categoryNameExists)

	if err != nil {
		return fmt.Errorf("%s: category name check failed: %w", op, err)
	}

	if !categoryNameExists {
		return fmt.Errorf("%s: %w: category name '%d' not found",
			op, storage.ErrNoCategoriesFound, CatID)
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
