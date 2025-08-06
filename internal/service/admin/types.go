package admin

import (
	"context"
	"errors"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
)

func (s *ServiceAdmin) AddType(ctx context.Context, req *admin.ContentType) (*admin.ContentType, error) {
	const op = "service.AddType"

	if req.Name == "" {
		logging.L(ctx).Error("empty required field", "type_content", req.Name, "op", op)
		return nil, admin.ErrTypeNameEmpty
	}

	res, err := s.db.AddType(ctx, req)
	if err != nil {
		logging.L(ctx).Error("failed to add type", sl.Err(err), "op", op)
		return nil, admin.ErrFailedSaveType
	}

	if s.cfg.Redis.Enable == true {
		go s.removeCache(ctx, op)
	}

	return res, nil
}

func (s *ServiceAdmin) GetType(ctx context.Context, id int64) (*admin.ContentType, error) {
	const op = "service.GetType"

	if id <= 0 {
		logging.L(ctx).Error("invalid type ID", "id", id, "op", op)
		return nil, admin.ErrTypeInvalid
	}

	res, err := s.db.GetType(ctx, id)
	if err != nil {
		if errors.Is(err, admin.ErrTypeNotFound) {
			logging.L(ctx).Warn("type not found", "id", id, "op", op)
			return nil, admin.ErrTypeNotFound
		}
		logging.L(ctx).Error("failed to get type", sl.Err(err), "op", op)
		return nil, admin.ErrFailedGetType
	}

	return res, nil
}

func (s *ServiceAdmin) GetTypes(ctx context.Context) ([]admin.ContentType, error) {
	const op = "service.GetTypes"

	res, err := s.db.GetTypes(ctx)
	if err != nil {
		if errors.Is(err, admin.ErrTypeNotFound) {
			logging.L(ctx).Warn("no types found", "op", op)
			return nil, admin.ErrTypeNotFound
		}
		logging.L(ctx).Error("failed to get types", sl.Err(err), "op", op)
		return nil, admin.ErrFailedGetType
	}

	return res, nil
}

func (s *ServiceAdmin) UpdateType(ctx context.Context, req *admin.ContentType) error {
	const op = "service.UpdateType"

	if req.ID <= 0 {
		logging.L(ctx).Error("invalid type ID", "id", req.ID, "op", op)
		return admin.ErrTypeInvalid
	}

	if req.Name == "" {
		logging.L(ctx).Error("empty required field", "type_content", req.Name, "op", op)
		return admin.ErrTypeNameEmpty
	}

	err := s.db.UpdateType(ctx, req)
	if err != nil {
		if errors.Is(err, admin.ErrTypeNotFound) {
			logging.L(ctx).Warn("type not found", "id", req.ID, "op", op)
			return admin.ErrTypeNotFound
		}
		logging.L(ctx).Error("failed to update type", sl.Err(err), "op", op)
		return admin.ErrFailedSaveType
	}

	if s.cfg.Redis.Enable == true {
		go s.removeCache(ctx, op)
	}

	return nil
}

func (s *ServiceAdmin) DeleteType(ctx context.Context, id int64) error {
	const op = "service.DeleteType"

	if id <= 0 {
		logging.L(ctx).Error("invalid type ID", "id", id, "op", op)
		return admin.ErrTypeInvalid
	}

	err := s.db.DeleteType(ctx, id)
	if err != nil {
		if errors.Is(err, admin.ErrTypeNotFound) {
			logging.L(ctx).Warn("type not found", "id", id, "op", op)
			return admin.ErrTypeNotFound
		}
		logging.L(ctx).Error("failed to delete type", sl.Err(err), "op", op)
		return admin.ErrFailedSaveType
	}

	if s.cfg.Redis.Enable == true {
		go s.removeCache(ctx, op)
	}

	return nil
}
