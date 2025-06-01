package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"time"
)

// AddType добавляет новый тип
func (s *Storage) AddType(ctx context.Context, req admResponse.TypeRequest) (admResponse.TypeResponse, error) {
	const op = "storage.postgres.AddType"

	query := `
        INSERT INTO content_types (name, deleted)
        VALUES ($1, FALSE)
        RETURNING id, name, created_at
    `

	var response admResponse.TypeResponse
	var createdAt time.Time

	err := s.db.QueryRowContext(ctx, query, req.Name).Scan(
		&response.ID,
		&response.Name,
		&createdAt,
	)

	if err != nil {
		return admResponse.TypeResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	// Форматируем дату
	response.DateCreated = createdAt.Format("02.01.2006")

	return response, nil
}

// GetType возвращает тип по ID
func (s *Storage) GetType(ctx context.Context, id int64) (admResponse.TypeResponse, error) {
	const op = "storage.postgres.GetType"

	query := `
        SELECT id, name, created_at
        FROM content_types
        WHERE id = $1 AND deleted IS NOT TRUE
    `

	var contentType admResponse.TypeResponse
	var createdAt time.Time

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&contentType.ID,
		&contentType.Name,
		&createdAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return admResponse.TypeResponse{}, sql.ErrNoRows
		}
		return admResponse.TypeResponse{}, fmt.Errorf("%s: %w", op, err)
	}

	// Форматируем дату
	contentType.DateCreated = createdAt.Format("02.01.2006")

	return contentType, nil
}

// GetTypes возвращает все типы
func (s *Storage) GetTypes(ctx context.Context) ([]admResponse.TypeResponse, error) {
	const op = "storage.postgres.GetTypes"

	query := `
        SELECT id, name, created_at
        FROM content_types
        WHERE deleted IS NOT TRUE
        ORDER BY id
    `

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var types []admResponse.TypeResponse
	for rows.Next() {
		var t admResponse.TypeResponse
		var createdAt time.Time

		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		// Форматируем дату
		t.DateCreated = createdAt.Format("02.01.2006")
		types = append(types, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return types, nil
}

// UpdateType обновляет данные типа
func (s *Storage) UpdateType(ctx context.Context, id int64, req admResponse.TypeRequest) error {
	const op = "storage.postgres.UpdateType"

	query := `
		UPDATE content_types
		SET name = $1
		WHERE id = $2 AND deleted IS NOT TRUE
	`

	result, err := s.db.ExecContext(ctx, query, req.Name, id)
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

// DeleteType помечает тип как удаленный
func (s *Storage) DeleteType(ctx context.Context, id int64) error {
	const op = "storage.postgres.DeleteType"

	query := `
		UPDATE content_types
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
