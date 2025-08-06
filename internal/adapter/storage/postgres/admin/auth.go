package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
)

func (s *Storage) GetAdminUser(ctx context.Context, login, passwordHash string) (*admin.Users, error) {
	const op = "storage.postgres.admin.GetAdminUser"

	query := `
        SELECT id, username, password, admin 
        FROM accounts 
        WHERE username = $1 AND password = $2 AND admin = TRUE AND deleted IS NOT TRUE
    `

	var user admin.Users
	err := s.db.QueryRow(ctx, query, login, passwordHash).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.IsAdmin,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, admin.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}
