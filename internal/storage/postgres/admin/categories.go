package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"time"
)

// AddCategory добавляет новую категорию
func (s *Storage) AddCategory(ctx context.Context, req *admResponse.CategoryRequest) (*admResponse.CategoryResponse, error) {
	const op = "storage.postgres.AddCategory"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Добавляем категорию
	var category admResponse.CategoryResponse
	var createdAt time.Time

	err = tx.QueryRowContext(ctx, `
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
		_, err = tx.ExecContext(ctx, `
			INSERT INTO category_content_types (category_id, content_type_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, category.ID, typeID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	// Получаем информацию о связанных типах
	rows, err := tx.QueryContext(ctx, `
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
		var t admResponse.TypeResponse
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		category.Types = append(category.Types, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &category, nil
}

// GetCategory возвращает категорию по ID
func (s *Storage) GetCategory(ctx context.Context, id int64) (*admResponse.CategoryResponse, error) {
	const op = "storage.postgres.GetCategory"

	var category admResponse.CategoryResponse
	var createdAt time.Time

	err := s.db.QueryRowContext(ctx, `
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	category.DateCreated = createdAt.Format("02.01.2006")
	category.Types = []admResponse.TypeResponse{}

	rows, err := s.db.QueryContext(ctx, `
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
		var t admResponse.TypeResponse

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
func (s *Storage) GetCategories(ctx context.Context) (*[]admResponse.CategoryResponse, error) {
	const op = "storage.postgres.GetCategories"

	// Сначала получаем все категории
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, img_url, created_at
		FROM categories
		WHERE deleted IS NOT TRUE
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var categories []admResponse.CategoryResponse
	for rows.Next() {
		var category admResponse.CategoryResponse
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
		category.Types = []admResponse.TypeResponse{}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Для каждой категории получаем связанные типы
	for i := range categories {
		rows, err := s.db.QueryContext(ctx, `
			SELECT ct.id, ct.name
			FROM content_types ct
			JOIN category_content_types cct ON ct.id = cct.content_type_id
			WHERE cct.category_id = $1
		`, categories[i].ID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		for rows.Next() {
			var t admResponse.TypeResponse
			if err := rows.Scan(&t.ID, &t.Name); err != nil {
				rows.Close()
				return nil, fmt.Errorf("%s: %w", op, err)
			}

			categories[i].Types = append(categories[i].Types, t)
		}

		if err = rows.Close(); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return &categories, nil
}

// UpdateCategory обновляет данные категории
func (s *Storage) UpdateCategory(ctx context.Context, id int64, req *admResponse.CategoryRequest) error {
	const op = "storage.postgres.UpdateCategory"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Обновляем основную информацию о категории
	result, err := tx.ExecContext(ctx, `
		UPDATE categories
		SET name = $1, img_url = $2
		WHERE id = $3 AND deleted IS NOT TRUE
	`, req.Name, req.ImgURL, id)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		err = sql.ErrNoRows
		return fmt.Errorf("%s: %w", op, err)
	}

	// Удаляем старые связи с типами
	_, err = tx.ExecContext(ctx, `
		DELETE FROM category_content_types
		WHERE category_id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Добавляем новые связи с типами
	for _, typeID := range req.TypeIDs {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO category_content_types (category_id, content_type_id)
			VALUES ($1, $2)
		`, id, typeID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// DeleteCategory помечает категорию как удаленную
func (s *Storage) DeleteCategory(ctx context.Context, id int64) error {
	const op = "storage.postgres.DeleteCategory"

	query := `
		UPDATE categories
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
