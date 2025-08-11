package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/internal/entities/api"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/middleware/metrics"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	rdb *redis.Client
	cfg *config.Config
}

// NewStorage создает новый экземпляр хранилища redis с предоставленным клиентом redis
func NewStorage(client *redis.Client, cfg *config.Config) *Storage {
	return &Storage{
		rdb: client,
		cfg: cfg,
	}
}

func InitRedis(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return rdb, nil
}

// GetCategories получает категории из кэша redis
func (s *Storage) GetCategories(ctx context.Context, typeID int64) ([]api.Category, error) {
	const op = "storage.redis.GetCategories"

	cacheKey := fmt.Sprintf("categories:%d", typeID)

	data, err := s.rdb.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Ключ не найден - это нормально, не считаем ошибкой
			return nil, err
		}
		return nil, fmt.Errorf("%s: failed to get from redis: %w", op, err)
	}

	var categories []api.Category
	if err = json.Unmarshal(data, &categories); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal categories: %w", op, err)
	}

	categories[0].DataSource = metrics.SourceRedis

	return categories, nil
}

// SetCategories сохраняет категории в кэш redis
func (s *Storage) SetCategories(ctx context.Context, typeID int64, categories []api.Category) error {
	const op = "storage.redis.SetCategories"

	cacheKey := fmt.Sprintf("categories:%d", typeID)

	data, err := json.Marshal(categories)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal categories: %w", op, err)
	}

	if err = s.rdb.Set(ctx, cacheKey, data, s.cfg.Redis.CacheTTL).Err(); err != nil {
		return fmt.Errorf("%s: failed to set redis key: %w", op, err)
	}

	return nil
}

// GetAccount получает данные аккаунта из кэша redis
func (s *Storage) GetAccount(ctx context.Context, account *api.Account) (*api.Account, error) {
	const op = "storage.redis.GetAccount"

	cacheKey := fmt.Sprintf("account:%s", account.Username)
	data, err := s.rdb.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, err
		}
		return nil, fmt.Errorf("%s: failed to get from redis: %w", op, err)
	}

	if err = json.Unmarshal(data, &account.ContentType); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal account: %w", op, err)
	}

	account.DataSource = metrics.SourceRedis

	return account, nil
}

// SetAccount сохраняет данные аккаунта в кэш redis
func (s *Storage) SetAccount(ctx context.Context, account *api.Account) error {
	const op = "storage.redis.SetAccount"

	cacheKey := fmt.Sprintf("account:%s", account.Username)
	data, err := json.Marshal(account.ContentType)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal account: %w", op, err)
	}

	if err = s.rdb.Set(ctx, cacheKey, data, s.cfg.Redis.CacheTTL).Err(); err != nil {
		return fmt.Errorf("%s: failed to set redis key: %w", op, err)
	}

	return nil
}

// GetVideo получает данные видео из кэша redis
func (s *Storage) GetVideo(ctx context.Context, videoID int64) (*api.Video, error) {
	const op = "storage.redis.GetVideo"

	cacheKey := fmt.Sprintf("video:%d", videoID)
	data, err := s.rdb.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, err
		}
		return nil, fmt.Errorf("%s: failed to get from redis: %w", op, err)
	}

	var video api.Video
	if err = json.Unmarshal(data, &video); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal video: %w", op, err)
	}

	video.DataSource = metrics.SourceRedis

	return &video, nil
}

// SetVideo сохраняет данные видео в кэш redis
func (s *Storage) SetVideo(ctx context.Context, videoID int64, video *api.Video) error {
	const op = "storage.redis.SetVideo"

	cacheKey := fmt.Sprintf("video:%d", videoID)
	data, err := json.Marshal(video)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal video: %w", op, err)
	}

	if err = s.rdb.Set(ctx, cacheKey, data, s.cfg.Redis.CacheTTL).Err(); err != nil {
		return fmt.Errorf("%s: failed to set redis key: %w", op, err)
	}

	return nil
}

// GetVideosByCategoryAndType получает видео из кэша redis
func (s *Storage) GetVideosByCategoryAndType(ctx context.Context, typeID, catID int64) ([]api.Video, error) {
	const op = "storage.redis.GetVideosByCategoryAndType"

	cacheKey := fmt.Sprintf("videos:%d:%d", typeID, catID)
	data, err := s.rdb.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, err
		}
		return nil, fmt.Errorf("%s: failed to get from redis: %w", op, err)
	}

	var videos []api.Video
	if err := json.Unmarshal(data, &videos); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal videos: %w", op, err)
	}

	videos[0].DataSource = metrics.SourceRedis

	return videos, nil
}

// SetVideosByCategoryAndType сохраняет видео в кэш redis
func (s *Storage) SetVideosByCategoryAndType(ctx context.Context, typeID, catID int64, videos []api.Video) error {
	const op = "storage.redis.SetVideosByCategoryAndType"

	cacheKey := fmt.Sprintf("videos:%d:%d", typeID, catID)
	data, err := json.Marshal(videos)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal videos: %w", op, err)
	}

	if err := s.rdb.Set(ctx, cacheKey, data, s.cfg.Redis.CacheTTL).Err(); err != nil {
		return fmt.Errorf("%s: failed to set redis key: %w", op, err)
	}

	return nil
}

// InvalidateCacheByPattern удаляет все ключи из кэша redis, соответствующие указанному шаблону
func (s *Storage) InvalidateCacheByPattern(ctx context.Context, pattern string) error {
	const op = "storage.redis.InvalidateCacheByPattern"

	var cursor uint64
	var totalDeleted int64

	for {
		var keys []string
		var err error
		keys, cursor, err = s.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("%s: failed to scan redis keys: %w", op, err)
		}

		if len(keys) > 0 {
			deleted, err := s.rdb.Del(ctx, keys...).Result()
			if err != nil {
				return fmt.Errorf("%s: failed to delete redis keys: %w", op, err)
			}
			totalDeleted += deleted
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}

// InvalidateAllCache удаляет весь кэш из redis
func (s *Storage) InvalidateAllCache(ctx context.Context) error {
	return s.InvalidateCacheByPattern(ctx, "*")
}

// InvalidateVideosCache удаляет весь кэш видео
func (s *Storage) InvalidateVideosCache(ctx context.Context) error {
	return s.InvalidateCacheByPattern(ctx, "videos:*")
}

// InvalidateCategoriesCache удаляет весь кэш категорий
func (s *Storage) InvalidateCategoriesCache(ctx context.Context) error {
	return s.InvalidateCacheByPattern(ctx, "categories:*")
}

// InvalidateAccountsCache удаляет весь кэш аккаунтов
func (s *Storage) InvalidateAccountsCache(ctx context.Context) error {
	return s.InvalidateCacheByPattern(ctx, "account:*")
}

func (s *Storage) HealthCheck(ctx context.Context) error {
	const op = "storage.redis.HealthCheck"

	if err := s.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("%s: failed to ping redis: %w", op, err)
	}

	return nil
}
