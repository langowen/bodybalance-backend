package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/redis/go-redis/v9"
	"github.com/theartofdevel/logging"
	"time"
)

type Storage struct {
	redis *redis.Client
	cfg   *config.Config
}

func New(cfg *config.Config) (*Storage, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host,     // Адрес Redis сервера
		Password: cfg.Redis.Password, // Пароль, если есть
		DB:       cfg.Redis.DB,       // Номер базы данных
	})

	// Проверка подключения
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	storage := &Storage{
		redis: redisClient,
		cfg:   cfg,
	}
	return storage, nil
}

// GetCategoriesWithFallback пытается получить категории из кэша, а если их нет - вызывает fallback функцию
func (s *Storage) GetCategoriesWithFallback(
	ctx context.Context,
	contentType string,
	ttl time.Duration,
	fallback func() ([]response.CategoryResponse, error),
) ([]response.CategoryResponse, error) {
	const op = "storage.redis.GetCategoriesWithFallback"

	// Пытаемся получить из кэша
	categories, err := s.GetCategories(ctx, contentType)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Если нашли в кэше - возвращаем
	if categories != nil {
		return categories, nil
	}

	// Если нет в кэше - вызываем fallback функцию
	categories, err = fallback()
	if err != nil {
		return nil, fmt.Errorf("%s: fallback failed: %w", op, err)
	}

	// Сохраняем результат в кэш
	if err := s.SetCategories(ctx, contentType, categories, ttl); err != nil {
		logging.L(ctx).Error("Failed to set fallback categories", sl.Err(err))
	}

	return categories, nil
}

// GetCategories получает категории из кэша Redis
func (s *Storage) GetCategories(ctx context.Context, contentType string) ([]response.CategoryResponse, error) {
	const op = "storage.redis.GetCategories"

	cacheKey := fmt.Sprintf("categories:%s", contentType)

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
func (s *Storage) SetCategories(ctx context.Context, contentType string, categories []response.CategoryResponse, ttl time.Duration) error {
	const op = "storage.redis.SetCategories"

	cacheKey := fmt.Sprintf("categories:%s", contentType)

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
func (s *Storage) DeleteCategories(ctx context.Context, contentType string) error {
	const op = "storage.redis.DeleteCategories"

	cacheKey := fmt.Sprintf("categories:%s", contentType)

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
func (s *Storage) GetVideo(ctx context.Context, videoID string) (*response.VideoResponse, error) {
	const op = "storage.redis.GetVideo"

	cacheKey := fmt.Sprintf("video:%s", videoID)
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
func (s *Storage) SetVideo(ctx context.Context, videoID string, video *response.VideoResponse, ttl time.Duration) error {
	const op = "storage.redis.SetVideo"

	cacheKey := fmt.Sprintf("video:%s", videoID)
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
func (s *Storage) DeleteVideo(ctx context.Context, videoID string) error {
	const op = "storage.redis.DeleteVideo"

	cacheKey := fmt.Sprintf("video:%s", videoID)
	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		return fmt.Errorf("%s: failed to delete redis key: %w", op, err)
	}

	return nil
}

// GetVideosByCategoryAndType получает видео из кэша Redis
func (s *Storage) GetVideosByCategoryAndType(ctx context.Context, contentType, category string) ([]response.VideoResponse, error) {
	const op = "storage.redis.GetVideosByCategoryAndType"

	cacheKey := fmt.Sprintf("videos:%s:%s", contentType, category)
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

	return videos, nil
}

// SetVideosByCategoryAndType сохраняет видео в кэш Redis
func (s *Storage) SetVideosByCategoryAndType(ctx context.Context, contentType, category string, videos []response.VideoResponse, ttl time.Duration) error {
	const op = "storage.redis.SetVideosByCategoryAndType"

	cacheKey := fmt.Sprintf("videos:%s:%s", contentType, category)
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
func (s *Storage) DeleteVideosByCategoryAndType(ctx context.Context, contentType, category string) error {
	const op = "storage.redis.DeleteVideosByCategoryAndType"

	cacheKey := fmt.Sprintf("videos:%s:%s", contentType, category)
	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		return fmt.Errorf("%s: failed to delete redis key: %w", op, err)
	}

	return nil
}
