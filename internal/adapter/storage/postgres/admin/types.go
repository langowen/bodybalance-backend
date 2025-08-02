package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin/dto"
	"time"
)

// AddType добавляет новый тип
func (s *Storage) AddType(ctx context.Context, req *dto.TypeRequest) (*dto.TypeResponse, error) {
	const op = "storage.postgres.AddType"

	query := `
        INSERT INTO content_types (name, deleted)
        VALUES ($1, FALSE)
        RETURNING id, name, created_at
    `

	var response dto.TypeResponse
	var createdAt time.Time

	err := s.db.QueryRow(ctx, query, req.Name).Scan(
		&response.ID,
		&response.Name,
		&createdAt,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Форматируем дату
	response.DateCreated = createdAt.Format("02.01.2006")

	return &response, nil
}

// GetType возвращает тип по ID
func (s *Storage) GetType(ctx context.Context, id int64) (*dto.TypeResponse, error) {
	const op = "storage.postgres.GetType"

	query := `
        SELECT id, name, created_at
        FROM content_types
        WHERE id = $1 AND deleted IS NOT TRUE
    `

	var contentType dto.TypeResponse
	var createdAt time.Time

	err := s.db.QueryRow(ctx, query, id).Scan(
		&contentType.ID,
		&contentType.Name,
		&createdAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Форматируем дату
	contentType.DateCreated = createdAt.Format("02.01.2006")

	return &contentType, nil
}

// GetTypes возвращает все типы
func (s *Storage) GetTypes(ctx context.Context) ([]dto.TypeResponse, error) {
	const op = "storage.postgres.GetTypes"

	query := `
        SELECT id, name, created_at
        FROM content_types
        WHERE deleted IS NOT TRUE
        ORDER BY id
    `

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var types []dto.TypeResponse
	for rows.Next() {
		var t dto.TypeResponse
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
func (s *Storage) UpdateType(ctx context.Context, id int64, req *dto.TypeRequest) error {
	const op = "storage.postgres.UpdateType"

	query := `
		UPDATE content_types
		SET name = $1
		WHERE id = $2 AND deleted IS NOT TRUE
	`

	commandTag, err := s.db.Exec(ctx, query, req.Name, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// DeleteType помечает тип как удаленный и удаляет все его связи с категориями
func (s *Storage) DeleteType(ctx context.Context, id int64) error {
	const op = "storage.postgres.DeleteType"

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
		WHERE content_type_id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("%s: failed to delete category relations: %w", op, err)
	}

	commandTag, err := tx.Exec(ctx, `
		UPDATE content_types
		SET deleted = TRUE
		WHERE id = $1 AND deleted IS NOT TRUE
	`, id)
	if err != nil {
		return fmt.Errorf("%s: failed to mark type as deleted: %w", op, err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: commit transaction failed: %w", op, err)
	}

	return nil
}
