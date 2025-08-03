package admin

import (
	"context"
	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
)

type ServiceAdmin struct {
	cfg   *config.Config
	db    AdmStorage
	redis CashStorage
}

func NewServiceAdmin(cfg *config.Config, storage AdmStorage, redis CashStorage) *ServiceAdmin {
	return &ServiceAdmin{
		cfg:   cfg,
		db:    storage,
		redis: redis,
	}
}

// removeCache удаляет записи из кэша при обновлении или удаления данных
func (s *ServiceAdmin) removeCache(ctx context.Context, op string) {

	err := s.redis.InvalidateVideosCache(ctx)
	if err != nil {
		logging.L(ctx).Warn("failed to invalidate videos cache", "op", op, sl.Err(err))
	}

	err = s.redis.InvalidateCategoriesCache(ctx)
	if err != nil {
		logging.L(ctx).Warn("failed to invalidate category cache", "op", op, sl.Err(err))
	}

	err = s.redis.InvalidateAccountsCache(ctx)
	if err != nil {
		logging.L(ctx).Warn("failed to invalidate account cache", "op", op, sl.Err(err))
	}

}
