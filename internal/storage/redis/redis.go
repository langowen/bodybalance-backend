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

	storage := &Storage{
		redis: redisClient.(*redis.Client),
		cfg:   cfg,
	}
	return storage, nil
}

// NewStorage создает новый экземпляр хранилища Redis с предоставленным клиентом Redis
// Используется для тестирования с помощью redismock.ClientMock
func NewStorage(client redis.UniversalClient) *Storage {
	return &Storage{
		redis: client.(*redis.Client),
		cfg:   &config.Config{}, // Пустой конфиг для тестирования
	}
}

// GetCategories получает категории из кэша Redis
func (s *Storage) GetCategories(ctx context.Context, typeID int64) (*[]response.CategoryResponse, error) {
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

	return &categories, nil
}

// SetCategories сохраняет категории в кэш Redis
func (s *Storage) SetCategories(ctx context.Context, typeID int64, categories *[]response.CategoryResponse, ttl time.Duration) error {
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

// DeleteCategories удаляет категории из кэша Redis
func (s *Storage) DeleteCategories(ctx context.Context, typeID int64) error {
	const op = "storage.redis.DeleteCategories"

	cacheKey := fmt.Sprintf("categories:%d", typeID)

	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		return fmt.Errorf("%s: failed to delete redis key: %w", op, err)
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

// DeleteAccount удаляет данные аккаунта из кэша Redis
func (s *Storage) DeleteAccount(ctx context.Context, username string) error {
	const op = "storage.redis.DeleteAccount"

	cacheKey := fmt.Sprintf("account:%s", username)
	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		return fmt.Errorf("%s: failed to delete redis key: %w", op, err)
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

// DeleteVideo удаляет данные видео из кэша Redis
func (s *Storage) DeleteVideo(ctx context.Context, videoID int64) error {
	const op = "storage.redis.DeleteVideo"

	cacheKey := fmt.Sprintf("video:%d", videoID)
	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		return fmt.Errorf("%s: failed to delete redis key: %w", op, err)
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
func (s *Storage) SetVideosByCategoryAndType(ctx context.Context, typeID, catID int64, videos *[]response.VideoResponse, ttl time.Duration) error {
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

// DeleteVideosByCategoryAndType удаляет видео из кэша Redis
func (s *Storage) DeleteVideosByCategoryAndType(ctx context.Context, typeID, catID int64) error {
	const op = "storage.redis.DeleteVideosByCategoryAndType"

	cacheKey := fmt.Sprintf("videos:%d:%d", typeID, catID)
	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		return fmt.Errorf("%s: failed to delete redis key: %w", op, err)
	}

	return nil
}
