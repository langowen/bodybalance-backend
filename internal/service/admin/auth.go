package admin

import (
	"context"
	"errors"

	"github.com/langowen/bodybalance-backend/internal/entities/admin"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
)

func (s *ServiceAdmin) Signing(ctx context.Context, login, password string) (*admin.Users, error) {
	const op = "service.admin.Signing"

	if login == "" {
		logging.L(ctx).Error("empty login", "op", op)

		return nil, admin.ErrEmptyUsername
	}

	if password == "" {
		logging.L(ctx).Error("empty password", "op", op)
		return nil, admin.ErrEmptyPassword
	}

	user, err := s.db.GetAdminUser(ctx, login, password)
	if err != nil {
		if errors.Is(err, admin.ErrUserNotFound) {
			logging.L(ctx).Warn("invalid login credentials", "op", op)
			return nil, err
		}
		logging.L(ctx).Error("failed to get user", "op", op, sl.Err(err))
		return nil, err
	}

	if !user.IsAdmin {
		logging.L(ctx).Warn("user is not admin", "op", op)
		return nil, admin.ErrUserNotAdmin
	}

	return user, nil
}
