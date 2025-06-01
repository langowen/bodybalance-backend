package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	err := s.db.QueryRowContext(ctx, query, login, passwordHash).Scan(
		&user.Username,
		&user.Password,
		&user.IsAdmin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}
