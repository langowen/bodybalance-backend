package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
)

type AdmUser struct {
	Username string
	Password string // Хэш пароля
	IsAdmin  bool
}

func (s *Storage) GetAdminUser(ctx context.Context, login, passwordHash string) (*AdmUser, error) {
	const op = "storage.postgres.admin.GetAdminUser"

	query := `
        SELECT username, password, admin 
        FROM accounts 
        WHERE username = $1 AND password = $2 AND admin = TRUE AND deleted IS NOT TRUE
    `

	var user AdmUser
	err := s.db.QueryRow(ctx, query, login, passwordHash).Scan(
		&user.Username,
		&user.Password,
		&user.IsAdmin,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}
