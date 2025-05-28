package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/storage"
	"strings"
)

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
