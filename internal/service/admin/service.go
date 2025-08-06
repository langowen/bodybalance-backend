package admin

import (
	"context"
	"regexp"
	"time"

	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
)

// validFilePattern паттерны для проверки правильности названия файлов
var validFilePattern = regexp.MustCompile(`^[a-zA-Z0-9_\-.]+\.[a-zA-Z0-9]+$`)

// suspiciousPatterns паттерны для проверки нет ли лишних символов и ссылок в данных
var suspiciousPatterns = []string{"://", "//", "../", "./", "\\", "?", "&", "=", "%"}

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
	ctxRedis, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.redis.InvalidateVideosCache(ctxRedis)
	if err != nil {
		logging.L(ctx).Warn("failed to invalidate videos cache", "service", op, sl.Err(err))
	}

	err = s.redis.InvalidateCategoriesCache(ctxRedis)
	if err != nil {
		logging.L(ctx).Warn("failed to invalidate category cache", "service", op, sl.Err(err))
	}

	err = s.redis.InvalidateAccountsCache(ctxRedis)
	if err != nil {
		logging.L(ctx).Warn("failed to invalidate account cache", "service", op, sl.Err(err))
	}

}
