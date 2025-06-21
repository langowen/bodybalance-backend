package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/redis/go-redis/v9"
	"time"
)

// Обертка над функцией создания клиента Redis для возможности мокирования в тестах
var redisNewClient = func(options *redis.Options) redis.UniversalClient {
	return redis.NewClient(options)
}

type Storage struct {
	redis *redis.Client
	cfg   *config.Config
}

// NewStorage создает новый экземпляр хранилища Redis с предоставленным клиентом Redis
func NewStorage(client redis.UniversalClient, cfg *config.Config) *Storage {
	return &Storage{
		redis: client.(*redis.Client),
		cfg:   cfg,
	}
}

func New(cfg *config.Config) (*Storage, error) {
	redisClient := redisNewClient(&redis.Options{
		Addr:     cfg.Redis.Host,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Проверка подключения
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	storage := NewStorage(redisClient, cfg)

	return storage, nil
}

// GetCategories получает категории из кэша Redis
func (s *Storage) GetCategories(ctx context.Context, typeID int64) ([]response.CategoryResponse, error) {
	const op = "storage.redis.GetCategories"

	cacheKey := fmt.Sprintf("categories:%d", typeID)

	// Пытаемся получить данные из Redis
	data, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Ключ не найден - это нормально, не считаем ошибкой
			return nil, nil
		}
		return nil, fmt.Errorf("%s: failed to get from redis: %w", op, err)
	}

	// Декодируем JSON данные
	var categories []response.CategoryResponse
	if err := json.Unmarshal(data, &categories); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal categories: %w", op, err)
	}

	return categories, nil
}

// SetCategories сохраняет категории в кэш Redis
func (s *Storage) SetCategories(ctx context.Context, typeID int64, categories []response.CategoryResponse, ttl time.Duration) error {
	const op = "storage.redis.SetCategories"

	cacheKey := fmt.Sprintf("categories:%d", typeID)

	// Сериализуем категории в JSON
	data, err := json.Marshal(categories)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal categories: %w", op, err)
	}

	// Сохраняем в Redis с указанным TTL
	if err := s.redis.Set(ctx, cacheKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("%s: failed to set redis key: %w", op, err)
	}

	return nil
}

// GetAccount получает данные аккаунта из кэша Redis
func (s *Storage) GetAccount(ctx context.Context, username string) (*response.AccountResponse, error) {
	const op = "storage.redis.GetAccount"

	cacheKey := fmt.Sprintf("account:%s", username)
	data, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: failed to get from redis: %w", op, err)
	}

	var account response.AccountResponse
	if err := json.Unmarshal(data, &account); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal account: %w", op, err)
	}

	return &account, nil
}

// SetAccount сохраняет данные аккаунта в кэш Redis
func (s *Storage) SetAccount(ctx context.Context, username string, account *response.AccountResponse, ttl time.Duration) error {
	const op = "storage.redis.SetAccount"

	cacheKey := fmt.Sprintf("account:%s", username)
	data, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal account: %w", op, err)
	}

	if err := s.redis.Set(ctx, cacheKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("%s: failed to set redis key: %w", op, err)
	}

	return nil
}

// GetVideo получает данные видео из кэша Redis
func (s *Storage) GetVideo(ctx context.Context, videoID int64) (*response.VideoResponse, error) {
	const op = "storage.redis.GetVideo"

	cacheKey := fmt.Sprintf("video:%d", videoID)
	data, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: failed to get from redis: %w", op, err)
	}

	var video response.VideoResponse
	if err := json.Unmarshal(data, &video); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal video: %w", op, err)
	}

	return &video, nil
}

// SetVideo сохраняет данные видео в кэш Redis
func (s *Storage) SetVideo(ctx context.Context, videoID int64, video *response.VideoResponse, ttl time.Duration) error {
	const op = "storage.redis.SetVideo"

	cacheKey := fmt.Sprintf("video:%d", videoID)
	data, err := json.Marshal(video)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal video: %w", op, err)
	}

	if err := s.redis.Set(ctx, cacheKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("%s: failed to set redis key: %w", op, err)
	}

	return nil
}

// GetVideosByCategoryAndType получает видео из кэша Redis
func (s *Storage) GetVideosByCategoryAndType(ctx context.Context, typeID, catID int64) (*[]response.VideoResponse, error) {
	const op = "storage.redis.GetVideosByCategoryAndType"

	cacheKey := fmt.Sprintf("videos:%d:%d", typeID, catID)
	data, err := s.redis.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: failed to get from redis: %w", op, err)
	}

	var videos []response.VideoResponse
	if err := json.Unmarshal(data, &videos); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal videos: %w", op, err)
	}

	return &videos, nil
}

// SetVideosByCategoryAndType сохраняет видео в кэш Redis
func (s *Storage) SetVideosByCategoryAndType(ctx context.Context, typeID, catID int64, videos []response.VideoResponse, ttl time.Duration) error {
	const op = "storage.redis.SetVideosByCategoryAndType"

	cacheKey := fmt.Sprintf("videos:%d:%d", typeID, catID)
	data, err := json.Marshal(videos)
	if err != nil {
		return fmt.Errorf("%s: failed to marshal videos: %w", op, err)
	}

	if err := s.redis.Set(ctx, cacheKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("%s: failed to set redis key: %w", op, err)
	}

	return nil
}

// InvalidateCacheByPattern удаляет все ключи из кэша Redis, соответствующие указанному шаблону
func (s *Storage) InvalidateCacheByPattern(ctx context.Context, pattern string) error {
	const op = "storage.redis.InvalidateCacheByPattern"

	var cursor uint64
	var totalDeleted int64

	for {
		var keys []string
		var err error
		keys, cursor, err = s.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("%s: failed to scan redis keys: %w", op, err)
		}

		if len(keys) > 0 {
			deleted, err := s.redis.Del(ctx, keys...).Result()
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

// InvalidateAllCache удаляет весь кэш из Redis
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
