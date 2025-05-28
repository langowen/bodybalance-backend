package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
)

// AddVideo добавляет новое видео в БД
func (s *Storage) AddVideo(ctx context.Context, video admResponse.VideoRequest) (int, error) {
	const op = "storage.postgres.AddVideo"

	query := `
		INSERT INTO videos (url, name, description, img_url, deleted)
		VALUES ($1, $2, $3, $4, FALSE)
		RETURNING id
	`

	var id int
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

// GetVideo возвращает видео по ID (только не удаленные)
func (s *Storage) GetVideo(ctx context.Context, id int64) (admResponse.VideoResponse, error) {
	const op = "storage.postgres.GetVideo"

	query := `
		SELECT id, url, name, description, img_url
		FROM videos
		WHERE id = $1 AND deleted IS NOT TRUE
	`

	var video admResponse.VideoResponse
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&video.ID,
		&video.URL,
		&video.Name,
		&video.Description,
		&video.ImgURL,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return admResponse.VideoResponse{}, sql.ErrNoRows
		}
		return admResponse.VideoResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	return video, nil
}

// GetVideos возвращает все не удаленные видео
func (s *Storage) GetVideos(ctx context.Context) ([]admResponse.VideoResponse, error) {
	const op = "storage.postgres.GetVideos"

	query := `
		SELECT id, url, name, description, img_url
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
		if err := rows.Scan(
			&video.ID,
			&video.URL,
			&video.Name,
			&video.Description,
			&video.ImgURL,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
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
