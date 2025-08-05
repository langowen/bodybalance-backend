package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
	"time"
)

// AddUser добавляет нового пользователя
func (s *Storage) AddUser(ctx context.Context, req *admin.Users) (*admin.Users, error) {
	const op = "storage.postgres.AddUser"

	query := `
        INSERT INTO accounts (username, content_type_id, admin, password, deleted)
        VALUES ($1, $2, $3, $4, FALSE)
        RETURNING id, username, content_type_id, 
                  (SELECT name FROM content_types WHERE id = $2),
                  admin, created_at
    `

	var user admin.Users
	var createdAt time.Time

	err := s.db.QueryRow(ctx, query,
		req.Username,
		req.ContentType.ID,
		req.IsAdmin,
		req.Password,
	).Scan(
		&user.ID,
		&user.Username,
		&user.ContentType.ID,
		&user.ContentType.Name,
		&user.IsAdmin,
		&createdAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, admin.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.DateCreated = createdAt.Format("02.01.2006")
	return &user, nil
}

// GetUser возвращает пользователя по ID
func (s *Storage) GetUser(ctx context.Context, id int64) (*admin.Users, error) {
	const op = "storage.postgres.GetUser"

	query := `
		SELECT a.id, a.username, a.content_type_id, 
		       ct.name, a.admin, a.created_at
		FROM accounts a
		LEFT JOIN content_types ct ON a.content_type_id = ct.id
		WHERE a.id = $1 AND a.deleted IS NOT TRUE
	`

	var user admin.Users
	var createdAt time.Time

	err := s.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.ContentType.ID,
		&user.ContentType.Name,
		&user.IsAdmin,
		&createdAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, admin.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.DateCreated = createdAt.Format("02.01.2006")
	return &user, nil
}

// GetUsers возвращает всех пользователей
func (s *Storage) GetUsers(ctx context.Context) ([]admin.Users, error) {
	const op = "storage.postgres.GetUsers"

	query := `
		SELECT a.id, a.username, a.content_type_id, 
		       ct.name, a.admin, a.created_at
		FROM accounts a
		LEFT JOIN content_types ct ON a.content_type_id = ct.id
		WHERE a.deleted IS NOT TRUE
		ORDER BY a.id
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var users []admin.Users
	for rows.Next() {
		var user admin.Users
		var createdAt time.Time

		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.ContentType.ID,
			&user.ContentType.Name,
			&user.IsAdmin,
			&createdAt,
		); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, admin.ErrUserNotFound
			}
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
func (s *Storage) UpdateUser(ctx context.Context, req *admin.Users) error {
	const op = "storage.postgres.UpdateUser"

	query := `
		UPDATE accounts
		SET username = $1, content_type_id = $2, admin = $3, password = $4
		WHERE id = $5 AND deleted IS NOT TRUE
	`

	commandTag, err := s.db.Exec(ctx, query,
		req.Username,
		req.ContentType.ID,
		req.IsAdmin,
		req.Password,
		req.ID,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return admin.ErrUserAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
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

	commandTag, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if commandTag.RowsAffected() == 0 {
		return admin.ErrUserNotFound
	}

	return nil
}
