package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"time"
)

// AddVideo добавляет новое видео в БД
func (s *Storage) AddVideo(ctx context.Context, video admResponse.VideoRequest) (int64, error) {
	const op = "storage.postgres.AddVideo"

	query := `
		INSERT INTO videos (url, name, description, img_url, deleted)
		VALUES ($1, $2, $3, $4, FALSE)
		RETURNING id
	`

	var id int64
	err := s.db.QueryRowContext(ctx, query,
		video.URL,
		video.Name,
		video.Description,
		video.ImgURL,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// AddVideoCategories добавляет связи видео с категориями
func (s *Storage) AddVideoCategories(ctx context.Context, videoID int64, categoryIDs []int64) error {
	const op = "storage.postgres.AddVideoCategories"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO video_categories (video_id, category_id)
		VALUES ($1, $2)
		ON CONFLICT (video_id, category_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	for _, catID := range categoryIDs {
		_, err := stmt.ExecContext(ctx, videoID, catID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return tx.Commit()
}

// GetVideo возвращает видео по ID (только не удаленные)
func (s *Storage) GetVideo(ctx context.Context, id int64) (admResponse.VideoResponse, error) {
	const op = "storage.postgres.GetVideo"

	query := `
		SELECT id, url, name, description, img_url, created_at
		FROM videos
		WHERE id = $1 AND deleted IS NOT TRUE
	`

	var video admResponse.VideoResponse
	var createdAt time.Time

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&video.ID,
		&video.URL,
		&video.Name,
		&video.Description,
		&video.ImgURL,
		&createdAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return admResponse.VideoResponse{}, sql.ErrNoRows
		}
		return admResponse.VideoResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	video.DateCreated = createdAt.Format("02.01.2006")

	return video, nil
}

// GetVideoCategories возвращает категории видео
func (s *Storage) GetVideoCategories(ctx context.Context, videoID int64) ([]admResponse.CategoryResponse, error) {
	const op = "storage.postgres.GetVideoCategories"

	query := `
		SELECT c.id, c.name
		FROM categories c
		JOIN video_categories vc ON c.id = vc.category_id
		WHERE vc.video_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, videoID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var categories []admResponse.CategoryResponse
	for rows.Next() {
		var cat admResponse.CategoryResponse
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return categories, nil
}

// GetVideos возвращает все не удаленные видео
func (s *Storage) GetVideos(ctx context.Context) ([]admResponse.VideoResponse, error) {
	const op = "storage.postgres.GetVideos"

	query := `
		SELECT id, url, name, description, img_url, created_at
		FROM videos
		WHERE deleted IS NOT TRUE
		ORDER BY id
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var videos []admResponse.VideoResponse
	for rows.Next() {
		var video admResponse.VideoResponse
		var createdAt time.Time

		if err := rows.Scan(
			&video.ID,
			&video.URL,
			&video.Name,
			&video.Description,
			&video.ImgURL,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		video.DateCreated = createdAt.Format("02.01.2006")
		videos = append(videos, video)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return videos, nil
}

// UpdateVideo обновляет данные видео
func (s *Storage) UpdateVideo(ctx context.Context, id int64, video admResponse.VideoRequest) error {
	const op = "storage.postgres.UpdateVideo"

	query := `
		UPDATE videos
		SET url = $1, name = $2, description = $3, img_url = $4
		WHERE id = $5 AND deleted IS NOT TRUE
	`

	result, err := s.db.ExecContext(ctx, query,
		video.URL,
		video.Name,
		video.Description,
		video.ImgURL,
		id,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteVideoCategories удаляет все связи видео с категориями
func (s *Storage) DeleteVideoCategories(ctx context.Context, videoID int64) error {
	const op = "storage.postgres.DeleteVideoCategories"

	query := `
		DELETE FROM video_categories
		WHERE video_id = $1
	`

	_, err := s.db.ExecContext(ctx, query, videoID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// DeleteVideo помечает видео как удаленное
func (s *Storage) DeleteVideo(ctx context.Context, id int64) error {
	const op = "storage.postgres.DeleteVideo"

	query := `
		UPDATE videos
		SET deleted = TRUE
		WHERE id = $1 AND deleted IS NOT TRUE
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
