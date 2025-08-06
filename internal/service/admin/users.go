package admin

import (
	"context"
	"errors"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
)

func (s *ServiceAdmin) AddUser(ctx context.Context, req *admin.Users) (*admin.Users, error) {
	const op = "service.AddUser"

	err := validUser(req)
	if err != nil {
		if errors.Is(err, admin.ErrEmptyUsername) {
			logging.L(ctx).Warn("empty required username", "op", op)
			return nil, admin.ErrEmptyUsername
		}
		if errors.Is(err, admin.ErrTypeInvalid) {
			logging.L(ctx).Warn("empty required type ID", "op", op)
			return nil, admin.ErrTypeInvalid
		}
	}

	if req.Password == "" && req.IsAdmin == true {
		logging.L(ctx).Warn("empty required password", "op", op)
		return nil, admin.ErrNotFoundPassword
	}

	user, err := s.db.AddUser(ctx, req)
	if err != nil {
		if errors.Is(err, admin.ErrUserAlreadyExists) {
			logging.L(ctx).Warn("user already exists", "username", req.Username, "op", op)
			return nil, admin.ErrUserAlreadyExists
		}
		logging.L(ctx).Error("failed to add user", sl.Err(err), "op", op)
		return nil, admin.ErrFailedSaveUser
	}

	if s.cfg.Redis.Enable == true {
		go s.removeCache(ctx, op)
	}

	return user, nil
}

func (s *ServiceAdmin) GetUser(ctx context.Context, id int64) (*admin.Users, error) {
	const op = "service.GetUser"

	if id <= 0 {
		logging.L(ctx).Error("invalid user ID", "id", id, "op", op)
		return nil, admin.ErrUserInvalidID
	}

	user, err := s.db.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, admin.ErrUserNotFound) {
			logging.L(ctx).Warn("user not found", "id", id, "op", op)
			return nil, admin.ErrUserNotFound
		}
		logging.L(ctx).Error("failed to get user", sl.Err(err), "op", op)
		return nil, admin.ErrFailedGetUser
	}

	return user, nil
}

func (s *ServiceAdmin) GetUsers(ctx context.Context) ([]admin.Users, error) {
	const op = "service.GetUsers"

	users, err := s.db.GetUsers(ctx)
	if err != nil {
		if errors.Is(err, admin.ErrUserNotFound) {
			logging.L(ctx).Warn("no users found", "op", op)
			return nil, admin.ErrUserNotFound
		}
		logging.L(ctx).Error("failed to get users", sl.Err(err), "op", op)
		return nil, admin.ErrFailedGetUser
	}

	return users, nil
}

func (s *ServiceAdmin) UpdateUser(ctx context.Context, req *admin.Users) error {
	const op = "service.UpdateUser"

	if req.ID <= 0 {
		logging.L(ctx).Error("invalid user ID", "id", req.ID, "op", op)
		return admin.ErrUserInvalidID
	}

	err := validUser(req)
	if err != nil {
		if errors.Is(err, admin.ErrEmptyUsername) {
			logging.L(ctx).Warn("empty required username", "op", op)
			return admin.ErrEmptyUsername
		}
		if errors.Is(err, admin.ErrTypeInvalid) {
			logging.L(ctx).Warn("empty required type ID", "op", op)
			return admin.ErrTypeInvalid
		}
	}

	err = s.db.UpdateUser(ctx, req)
	if err != nil {
		if errors.Is(err, admin.ErrUserNotFound) {
			logging.L(ctx).Warn("user not found", "id", req.ID, "op", op)
			return admin.ErrUserNotFound
		}
		if errors.Is(err, admin.ErrUserAlreadyExists) {
			logging.L(ctx).Warn("user already exists", "username", req.Username, "op", op)
			return admin.ErrUserAlreadyExists
		}
		logging.L(ctx).Error("failed to update user", sl.Err(err), "op", op)
		return admin.ErrFailedSaveUser
	}

	if s.cfg.Redis.Enable == true {
		go s.removeCache(ctx, op)
	}

	return nil
}

func (s *ServiceAdmin) DeleteUser(ctx context.Context, id int64) error {
	const op = "service.DeleteUser"

	if id <= 0 {
		logging.L(ctx).Error("invalid user ID", "id", id, "op", op)
		return admin.ErrUserInvalidID
	}

	err := s.db.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, admin.ErrUserNotFound) {
			logging.L(ctx).Warn("user not found", "id", id, "op", op)
			return admin.ErrUserNotFound
		}
		logging.L(ctx).Error("failed to delete user", sl.Err(err), "op", op)
		return admin.ErrFailedDeleteUser
	}

	if s.cfg.Redis.Enable == true {
		go s.removeCache(ctx, op)
	}

	return nil
}

// validUser проверят входящие данные на валидность
func validUser(req *admin.Users) error {
	switch {
	case req.Username == "":
		return admin.ErrEmptyUsername
	case req.ContentType.ID == 0:
		return admin.ErrTypeInvalid
	}

	return nil
}
