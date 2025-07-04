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
func (s *Storage) AddVideo(ctx context.Context, video *admResponse.VideoRequest) (int64, error) {
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
func (s *Storage) GetVideo(ctx context.Context, id int64) (*admResponse.VideoResponse, error) {
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
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	video.DateCreated = createdAt.Format("02.01.2006")

	return &video, nil
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

		// Получаем типы контента для данной категории
		typesQuery := `
			SELECT ct.id, ct.name
			FROM content_types ct
			JOIN category_content_types cct ON ct.id = cct.content_type_id
			WHERE cct.category_id = $1
		`

		typesRows, err := s.db.QueryContext(ctx, typesQuery, cat.ID)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to get content types for category %d: %w", op, cat.ID, err)
		}

		var types []admResponse.TypeResponse
		for typesRows.Next() {
			var t admResponse.TypeResponse
			if err := typesRows.Scan(&t.ID, &t.Name); err != nil {
				typesRows.Close()
				return nil, fmt.Errorf("%s: failed to scan content type: %w", op, err)
			}
			types = append(types, t)
		}

		typesRows.Close()

		if err := typesRows.Err(); err != nil {
			return nil, fmt.Errorf("%s: error after scanning content types: %w", op, err)
		}

		// Устанавливаем типы для категории
		cat.Types = types

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
func (s *Storage) UpdateVideo(ctx context.Context, id int64, video *admResponse.VideoRequest) error {
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

// DeleteVideo помечает видео как удаленное и удаляет все его связи с категориями
func (s *Storage) DeleteVideo(ctx context.Context, id int64) error {
	const op = "storage.postgres.DeleteVideo"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: begin transaction failed: %w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, `
		DELETE FROM video_categories
		WHERE video_id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("%s: failed to delete category relations: %w", op, err)
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE videos
		SET deleted = TRUE
		WHERE id = $1 AND deleted IS NOT TRUE
	`, id)
	if err != nil {
		return fmt.Errorf("%s: failed to mark video as deleted: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: rows affected error: %w", op, err)
	}

	if rowsAffected == 0 {
		err = sql.ErrNoRows
		return sql.ErrNoRows
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit transaction failed: %w", op, err)
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
