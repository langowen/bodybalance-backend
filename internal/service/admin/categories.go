package admin

import (
	"context"
	"errors"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"strings"
)

func (s *ServiceAdmin) AddCategory(ctx context.Context, req *admin.Category) (*admin.Category, error) {
	const op = "service.AddCategory"

	err := s.validCat(req)
	if err != nil {
		logging.L(ctx).Error("invalid category data", "op", op, "category", req, sl.Err(err))
		return nil, err
	}

	category, err := s.db.AddCategory(ctx, req)
	if err != nil {
		logging.L(ctx).Error("failed to add category", "op", op, "category", category, sl.Err(err))
		return nil, err
	}

	if s.cfg.Redis.Enable == true {
		go s.removeCache(ctx, op)
	}

	return category, nil
}

func (s *ServiceAdmin) GetCategory(ctx context.Context, id int64) (*admin.Category, error) {
	const op = "service.GetCategory"

	category, err := s.db.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, admin.ErrCategoryNotFound) {
			logging.L(ctx).Warn("category not found", "op", op, "category_id", id, sl.Err(err))
			return nil, err
		}
		logging.L(ctx).Error("failed to get category", "op", op, "category_id", id, sl.Err(err))
		return nil, err
	}

	return category, nil
}

func (s *ServiceAdmin) GetCategories(ctx context.Context) ([]admin.Category, error) {
	const op = "service.GetCategories"

	categories, err := s.db.GetCategories(ctx)
	if err != nil {
		logging.L(ctx).Error("failed to get categories", "op", op, sl.Err(err))
		return nil, err
	}

	return categories, nil
}

func (s *ServiceAdmin) UpdateCategory(ctx context.Context, id int64, req *admin.Category) error {
	const op = "service.UpdateCategory"

	err := s.validCat(req)
	if err != nil {
		logging.L(ctx).Error("invalid category data", "op", op, "category_id", id, sl.Err(err))
		return err
	}

	err = s.db.UpdateCategory(ctx, id, req)
	if err != nil {
		if errors.Is(err, admin.ErrCategoryNotFound) {
			logging.L(ctx).Warn("category not found", "op", op, "category_id", id, sl.Err(err))
			return err
		}
		logging.L(ctx).Error("failed to update category", "op", op, "category_id", id, sl.Err(err))
		return err
	}

	if s.cfg.Redis.Enable == true {
		go s.removeCache(ctx, op)
	}

	return nil
}

func (s *ServiceAdmin) DeleteCategory(ctx context.Context, id int64) error {
	const op = "service.DeleteCategory"

	err := s.db.DeleteCategory(ctx, id)
	if err != nil {
		if errors.Is(err, admin.ErrCategoryNotFound) {
			logging.L(ctx).Warn("category not found", "op", op, "category_id", id, sl.Err(err))
			return err
		}
		logging.L(ctx).Error("failed to delete category", "op", op, "category_id", id, sl.Err(err))
		return err
	}

	if s.cfg.Redis.Enable == true {
		go s.removeCache(ctx, op)
	}

	return nil
}

// validCat проверят данные входящего запроса на их валидность
func (s *ServiceAdmin) validCat(category *admin.Category) error {
	switch {
	case category.Name == "":
		return admin.ErrEmptyName
	case category.ImgURL == "":
		return admin.ErrEmptyImgURL
	case len(category.ContentType) == 0 || category.ContentType[0].ID == 0:
		return admin.ErrEmptyTypeIDs
	case !validFilePattern.MatchString(category.ImgURL):
		return admin.ErrInvalidImgFormat
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(category.ImgURL, pattern) {
			return admin.ErrSuspiciousContent
		}
	}

	return nil
}
