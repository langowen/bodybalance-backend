package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
	"time"
)

// AddVideo добавляет новое видео в БД
func (s *Storage) AddVideo(ctx context.Context, video *admin.Video) (int64, error) {
	const op = "storage.postgres.AddVideo"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	// 1. Вставляем видео
	var videoID int64
	err = tx.QueryRow(ctx, `
        INSERT INTO videos (url, name, description, img_url, deleted)
        VALUES ($1, $2, $3, $4, FALSE)
        RETURNING id
    `,
		video.URL,
		video.Name,
		video.Description,
		video.ImgURL,
	).Scan(&videoID)

	if err != nil {
		return 0, fmt.Errorf("%s: failed to insert video: %w", op, err)
	}

	// 2. Добавляем связи с категориями
	if len(video.Categories) > 0 {
		for _, cat := range video.Categories {
			_, err := tx.Exec(ctx, `
                INSERT INTO video_categories (video_id, category_id)
                VALUES ($1, $2)
                ON CONFLICT (video_id, category_id) DO NOTHING
            `, videoID, cat.ID)

			if err != nil {
				return 0, fmt.Errorf("%s: failed to add category %d: %w", op, cat.ID, err)
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("%s: transaction commit failed: %w", op, err)
	}

	return videoID, nil
}

// GetVideo возвращает видео по ID (только не удаленные)
func (s *Storage) GetVideo(ctx context.Context, id int64) (*admin.Video, error) {
	const op = "storage.postgres.GetVideo"

	// Начинаем read-only транзакцию для обеспечения консистентности данных
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer tx.Rollback(ctx)

	// Сначала получаем данные видео
	videoQuery := `
        SELECT id, url, name, description, img_url, created_at
        FROM videos
        WHERE id = $1 AND deleted IS NOT TRUE
    `

	var video admin.Video
	var createdAt time.Time

	err = tx.QueryRow(ctx, videoQuery, id).Scan(
		&video.ID,
		&video.URL,
		&video.Name,
		&video.Description,
		&video.ImgURL,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, admin.ErrVideoNotFound
		}
		return nil, fmt.Errorf("%s: failed to get video: %w", op, err)
	}

	video.DateCreated = createdAt.Format("02.01.2006")
	video.Categories = make([]admin.Category, 0)

	// Теперь получаем категории для этого видео в той же транзакции
	categoryQuery := `
        SELECT c.id, c.name, c.img_url, c.created_at
        FROM video_categories vc
        INNER JOIN categories c ON c.id = vc.category_id
        WHERE vc.video_id = $1 AND c.deleted IS NOT TRUE
        ORDER BY c.name
    `

	categoryRows, err := tx.Query(ctx, categoryQuery, id)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get categories: %w", op, err)
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		var category admin.Category
		var catCreatedAt time.Time

		err := categoryRows.Scan(
			&category.ID,
			&category.Name,
			&category.ImgURL,
			&catCreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan category: %w", op, err)
		}

		category.CreatedAt = catCreatedAt.Format("02.01.2006")
		video.Categories = append(video.Categories, category)
	}

	if err := categoryRows.Err(); err != nil {
		return nil, fmt.Errorf("%s: category rows error: %w", op, err)
	}

	// Коммитим read-only транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	return &video, nil
}

// GetVideos возвращает все не удаленные видео с их категориями
func (s *Storage) GetVideos(ctx context.Context) ([]admin.Video, error) {
	const op = "storage.postgres.GetVideos"

	// Начинаем read-only транзакцию для обеспечения консистентности данных
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer tx.Rollback(ctx)

	// Сначала получаем все видео
	videoQuery := `
        SELECT id, url, name, description, img_url, created_at
        FROM videos
        WHERE deleted IS NOT TRUE
        ORDER BY created_at DESC, id
    `

	videoRows, err := tx.Query(ctx, videoQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to query videos: %w", op, err)
	}
	defer videoRows.Close()

	var videos []admin.Video
	var videoIDs []int64

	for videoRows.Next() {
		var video admin.Video
		var createdAt time.Time

		err := videoRows.Scan(
			&video.ID,
			&video.URL,
			&video.Name,
			&video.Description,
			&video.ImgURL,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan video: %w", op, err)
		}

		video.DateCreated = createdAt.Format("02.01.2006")
		video.Categories = make([]admin.Category, 0)
		videos = append(videos, video)
		videoIDs = append(videoIDs, video.ID)
	}

	if err := videoRows.Err(); err != nil {
		return nil, fmt.Errorf("%s: video rows error: %w", op, err)
	}

	if len(videos) == 0 {
		return nil, admin.ErrVideoNotFound
	}

	// Теперь получаем все категории для этих видео одним запросом в той же транзакции
	categoryQuery := `
        SELECT 
            vc.video_id,
            c.id, c.name, c.img_url, c.created_at
        FROM video_categories vc
        INNER JOIN categories c ON c.id = vc.category_id
        WHERE vc.video_id = ANY($1) AND c.deleted IS NOT TRUE
        ORDER BY vc.video_id, c.name
    `

	categoryRows, err := tx.Query(ctx, categoryQuery, videoIDs)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to query categories: %w", op, err)
	}
	defer categoryRows.Close()

	// Создаем map для быстрого поиска видео по ID
	videoMap := make(map[int64]*admin.Video, len(videos))
	for i := range videos {
		videoMap[videos[i].ID] = &videos[i]
	}

	// Добавляем категории к соответствующим видео
	for categoryRows.Next() {
		var videoID int64
		var category admin.Category
		var createdAt time.Time

		err := categoryRows.Scan(
			&videoID,
			&category.ID,
			&category.Name,
			&category.ImgURL,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan category: %w", op, err)
		}

		category.CreatedAt = createdAt.Format("02.01.2006")

		if video, exists := videoMap[videoID]; exists {
			video.Categories = append(video.Categories, category)
		}
	}

	if err := categoryRows.Err(); err != nil {
		return nil, fmt.Errorf("%s: category rows error: %w", op, err)
	}

	// Коммитим read-only транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	return videos, nil
}

// UpdateVideo обновляет данные видео и его связи с категориями
func (s *Storage) UpdateVideo(ctx context.Context, video *admin.Video) error {
	const op = "storage.postgres.UpdateVideo"

	// Начинаем транзакцию для атомарного обновления
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer tx.Rollback(ctx)

	// 1. Обновляем основные данные видео
	updateVideoQuery := `
		UPDATE videos
		SET url = $1, name = $2, description = $3, img_url = $4
		WHERE id = $5 AND deleted IS NOT TRUE
	`

	commandTag, err := tx.Exec(ctx, updateVideoQuery,
		video.URL,
		video.Name,
		video.Description,
		video.ImgURL,
		video.ID,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to update video: %w", op, err)
	}

	if commandTag.RowsAffected() == 0 {
		return admin.ErrVideoNotFound
	}

	// 2. Получаем существующие связи с категориями
	existingCategoriesQuery := `
		SELECT category_id 
		FROM video_categories 
		WHERE video_id = $1
	`

	rows, err := tx.Query(ctx, existingCategoriesQuery, video.ID)
	if err != nil {
		return fmt.Errorf("%s: failed to get existing categories: %w", op, err)
	}
	defer rows.Close()

	existingCategories := make(map[int64]bool)
	for rows.Next() {
		var categoryID int64
		if err := rows.Scan(&categoryID); err != nil {
			return fmt.Errorf("%s: failed to scan category ID: %w", op, err)
		}
		existingCategories[categoryID] = true
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("%s: rows error: %w", op, err)
	}

	// 3. Создаем мапу новых категорий
	newCategories := make(map[int64]bool)
	for _, cat := range video.Categories {
		newCategories[cat.ID] = true
	}

	// 4. Удаляем категории, которых нет в новом списке
	categoriesToRemove := make([]int64, 0)
	for categoryID := range existingCategories {
		if !newCategories[categoryID] {
			categoriesToRemove = append(categoriesToRemove, categoryID)
		}
	}

	if len(categoriesToRemove) > 0 {
		deleteCategoriesQuery := `
			DELETE FROM video_categories 
			WHERE video_id = $1 AND category_id = ANY($2)
		`
		_, err := tx.Exec(ctx, deleteCategoriesQuery, video.ID, categoriesToRemove)
		if err != nil {
			return fmt.Errorf("%s: failed to remove categories: %w", op, err)
		}
	}

	// 5. Добавляем новые категории
	for _, cat := range video.Categories {
		if !existingCategories[cat.ID] {
			insertCategoryQuery := `
				INSERT INTO video_categories (video_id, category_id)
				VALUES ($1, $2)
				ON CONFLICT (video_id, category_id) DO NOTHING
			`
			_, err := tx.Exec(ctx, insertCategoryQuery, video.ID, cat.ID)
			if err != nil {
				return fmt.Errorf("%s: failed to add category %d: %w", op, cat.ID, err)
			}
		}
	}

	// 6. Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	return nil
}

// DeleteVideo помечает видео как удаленное и удаляет все его связи с категориями
func (s *Storage) DeleteVideo(ctx context.Context, id int64) error {
	const op = "storage.postgres.DeleteVideo"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: begin transaction failed: %w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, `
		DELETE FROM video_categories
		WHERE video_id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("%s: failed to delete category relations: %w", op, err)
	}

	commandTag, err := tx.Exec(ctx, `
		UPDATE videos
		SET deleted = TRUE
		WHERE id = $1 AND deleted IS NOT TRUE
	`, id)
	if err != nil {
		return fmt.Errorf("%s: failed to mark video as deleted: %w", op, err)
	}

	if commandTag.RowsAffected() == 0 {
		return admin.ErrVideoNotFound
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: commit transaction failed: %w", op, err)
	}

	return nil
}
