package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"strings"
	"time"
)

// AddUser добавляет нового пользователя
func (s *Storage) AddUser(ctx context.Context, req *admResponse.UserRequest) (*admResponse.UserResponse, error) {
	const op = "storage.postgres.AddUser"

	query := `
        INSERT INTO accounts (username, content_type_id, admin, password, deleted)
        VALUES ($1, $2, $3, $4, FALSE)
        RETURNING id, username, content_type_id, 
                  (SELECT name FROM content_types WHERE id = $2),
                  admin, created_at
    `

	var user admResponse.UserResponse
	var createdAt time.Time

	err := s.db.QueryRowContext(ctx, query,
		req.Username,
		req.ContentTypeID,
		req.Admin,
		req.Password,
	).Scan(
		&user.ID,
		&user.Username,
		&user.ContentTypeID,
		&user.ContentType,
		&user.Admin,
		&createdAt,
	)

	if err != nil {
		// Проверяем, является ли ошибка ошибкой дубликата
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, fmt.Errorf("%s: user already exists", op)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.DateCreated = createdAt.Format("02.01.2006")
	return &user, nil
}

// GetUser возвращает пользователя по ID
func (s *Storage) GetUser(ctx context.Context, id int64) (*admResponse.UserResponse, error) {
	const op = "storage.postgres.GetUser"

	query := `
		SELECT a.id, a.username, a.content_type_id, 
		       ct.name, a.admin, a.created_at
		FROM accounts a
		LEFT JOIN content_types ct ON a.content_type_id = ct.id
		WHERE a.id = $1 AND a.deleted IS NOT TRUE
	`

	var user admResponse.UserResponse
	var createdAt time.Time

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.ContentTypeID,
		&user.ContentType,
		&user.Admin,
		&createdAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.DateCreated = createdAt.Format("02.01.2006")
	return &user, nil
}

// GetUsers возвращает всех пользователей
func (s *Storage) GetUsers(ctx context.Context) ([]admResponse.UserResponse, error) {
	const op = "storage.postgres.GetUsers"

	query := `
		SELECT a.id, a.username, a.content_type_id, 
		       ct.name, a.admin, a.created_at
		FROM accounts a
		LEFT JOIN content_types ct ON a.content_type_id = ct.id
		WHERE a.deleted IS NOT TRUE
		ORDER BY a.id
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var users []admResponse.UserResponse
	for rows.Next() {
		var user admResponse.UserResponse
		var createdAt time.Time

		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.ContentTypeID,
			&user.ContentType,
			&user.Admin,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		user.DateCreated = createdAt.Format("02.01.2006")
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

// UpdateUser обновляет данные пользователя
func (s *Storage) UpdateUser(ctx context.Context, id int64, req *admResponse.UserRequest) error {
	const op = "storage.postgres.UpdateUser"

	query := `
		UPDATE accounts
		SET username = $1, content_type_id = $2, admin = $3, password = $4
		WHERE id = $5 AND deleted IS NOT TRUE
	`

	result, err := s.db.ExecContext(ctx, query,
		req.Username,
		req.ContentTypeID,
		req.Admin,
		req.Password,
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

// DeleteUser помечает пользователя как удаленного
func (s *Storage) DeleteUser(ctx context.Context, id int64) error {
	const op = "storage.postgres.DeleteUser"

	query := `
		UPDATE accounts
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
