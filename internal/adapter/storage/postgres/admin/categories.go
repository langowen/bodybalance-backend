package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin/dto"
	"time"
)

// AddCategory добавляет новую категорию
func (s *Storage) AddCategory(ctx context.Context, req *dto.CategoryRequest) (*dto.CategoryResponse, error) {
	const op = "storage.postgres.AddCategory"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Добавляем категорию
	var category dto.CategoryResponse
	var createdAt time.Time

	err = tx.QueryRow(ctx, `
		INSERT INTO categories (name, img_url, deleted)
		VALUES ($1, $2, FALSE)
		RETURNING id, name, img_url, created_at
	`, req.Name, req.ImgURL).Scan(
		&category.ID,
		&category.Name,
		&category.ImgURL,
		&createdAt,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	category.DateCreated = createdAt.Format("02.01.2006")

	// Добавляем связи с типами
	for _, typeID := range req.TypeIDs {
		_, err = tx.Exec(ctx, `
			INSERT INTO category_content_types (category_id, content_type_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, category.ID, typeID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	// Получаем информацию о связанных типах
	rows, err := tx.Query(ctx, `
		SELECT ct.id, ct.name
		FROM content_types ct
		JOIN category_content_types cct ON ct.id = cct.content_type_id
		WHERE cct.category_id = $1
	`, category.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var t dto.TypeResponse
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		category.Types = append(category.Types, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &category, nil
}

// GetCategory возвращает категорию по ID
func (s *Storage) GetCategory(ctx context.Context, id int64) (*dto.CategoryResponse, error) {
	const op = "storage.postgres.GetCategory"

	var category dto.CategoryResponse
	var createdAt time.Time

	err := s.db.QueryRow(ctx, `
		SELECT id, name, img_url, created_at
		FROM categories
		WHERE id = $1 AND deleted IS NOT TRUE
	`, id).Scan(
		&category.ID,
		&category.Name,
		&category.ImgURL,
		&createdAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	category.DateCreated = createdAt.Format("02.01.2006")
	category.Types = []dto.TypeResponse{}

	rows, err := s.db.Query(ctx, `
		SELECT ct.id, ct.name
		FROM content_types ct
		JOIN category_content_types cct ON ct.id = cct.content_type_id
		WHERE cct.category_id = $1
	`, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var t dto.TypeResponse

		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		category.Types = append(category.Types, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &category, nil
}

// GetCategories возвращает все категории
func (s *Storage) GetCategories(ctx context.Context) ([]dto.CategoryResponse, error) {
	const op = "storage.postgres.GetCategories"

	// Сначала получаем все категории
	rows, err := s.db.Query(ctx, `
		SELECT id, name, img_url, created_at
		FROM categories
		WHERE deleted IS NOT TRUE
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var categories []dto.CategoryResponse
	for rows.Next() {
		var category dto.CategoryResponse
		var createdAt time.Time

		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.ImgURL,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		category.DateCreated = createdAt.Format("02.01.2006")
		category.Types = []dto.TypeResponse{}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Для каждой категории получаем связанные типы
	for i := range categories {
		rows, err := s.db.Query(ctx, `
			SELECT ct.id, ct.name
			FROM content_types ct
			JOIN category_content_types cct ON ct.id = cct.content_type_id
			WHERE cct.category_id = $1
		`, categories[i].ID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		for rows.Next() {
			var t dto.TypeResponse
			if err := rows.Scan(&t.ID, &t.Name); err != nil {
				rows.Close()
				return nil, fmt.Errorf("%s: %w", op, err)
			}

			categories[i].Types = append(categories[i].Types, t)
		}

		rows.Close()
		if rows.Err() != nil {
			return nil, fmt.Errorf("%s: %w", op, rows.Err())
		}
	}

	return categories, nil
}

// UpdateCategory обновляет данные категории
func (s *Storage) UpdateCategory(ctx context.Context, id int64, req *dto.CategoryRequest) error {
	const op = "storage.postgres.UpdateCategory"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Обновляем основную информацию о категории
	commandTag, err := tx.Exec(ctx, `
		UPDATE categories
		SET name = $1, img_url = $2
		WHERE id = $3 AND deleted IS NOT TRUE
	`, req.Name, req.ImgURL, id)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	// Удаляем старые связи с типами
	_, err = tx.Exec(ctx, `
		DELETE FROM category_content_types
		WHERE category_id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Добавляем новые связи с типами
	for _, typeID := range req.TypeIDs {
		_, err = tx.Exec(ctx, `
			INSERT INTO category_content_types (category_id, content_type_id)
			VALUES ($1, $2)
		`, id, typeID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// DeleteCategory помечает категорию как удаленную и удаляет её связи с типами контента и видео
func (s *Storage) DeleteCategory(ctx context.Context, id int64) error {
	const op = "storage.postgres.DeleteCategory"

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
		DELETE FROM category_content_types
		WHERE category_id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("%s: failed to delete content type relations: %w", op, err)
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM video_categories
		WHERE category_id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("%s: failed to delete video relations: %w", op, err)
	}

	commandTag, err := tx.Exec(ctx, `
		UPDATE categories
		SET deleted = TRUE
		WHERE id = $1 AND deleted IS NOT TRUE
	`, id)
	if err != nil {
		return fmt.Errorf("%s: failed to mark category as deleted: %w", op, err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: commit transaction failed: %w", op, err)
	}

	return nil
}
