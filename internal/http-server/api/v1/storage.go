package v1

import (
	"context"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"time"
)

// ApiStorage определяет методы, которые нужны API v1 для работы с хранилищем.
type ApiStorage interface {
	GetVideosByCategoryAndType(ctx context.Context, contentType, categoryName string) ([]response.VideoResponse, error)
	GetCategories(ctx context.Context, contentType string) ([]response.CategoryResponse, error)
	CheckAccount(ctx context.Context, username string) (response.AccountResponse, error)
	GetVideo(ctx context.Context, videoID string) (response.VideoResponse, error)
}

// RedisApi интерфейс для работы с Redis хранилищем
type RedisApi interface {
	// GetCategoriesWithFallback пытается получить категории из кэша, а если их нет - вызывает fallback функцию
	GetCategoriesWithFallback(ctx context.Context, contentType string, ttl time.Duration, fallback func() ([]response.CategoryResponse, error)) ([]response.CategoryResponse, error)

	// GetCategories получает категории из кэша Redis
	GetCategories(ctx context.Context, contentType string) ([]response.CategoryResponse, error)

	// SetCategories сохраняет категории в кэш Redis
	SetCategories(ctx context.Context, contentType string, categories []response.CategoryResponse, ttl time.Duration) error

	// DeleteCategories удаляет категории из кэша Redis
	DeleteCategories(ctx context.Context, contentType string) error

	// GetAccount получает данные аккаунта из кэша Redis
	GetAccount(ctx context.Context, username string) (*response.AccountResponse, error)

	// SetAccount сохраняет данные аккаунта в кэш Redis
	SetAccount(ctx context.Context, username string, account *response.AccountResponse, ttl time.Duration) error

	// DeleteAccount удаляет данные аккаунта из кэша Redis
	DeleteAccount(ctx context.Context, username string) error

	// GetVideo получает данные видео из кэша Redis
	GetVideo(ctx context.Context, videoID string) (*response.VideoResponse, error)

	// SetVideo сохраняет данные видео в кэш Redis
	SetVideo(ctx context.Context, videoID string, video *response.VideoResponse, ttl time.Duration) error

	// DeleteVideo удаляет данные видео из кэша Redis
	DeleteVideo(ctx context.Context, videoID string) error

	// GetVideosByCategoryAndType получает видео из кэша Redis
	GetVideosByCategoryAndType(ctx context.Context, contentType, category string) ([]response.VideoResponse, error)

	// SetVideosByCategoryAndType сохраняет видео в кэш Redis
	SetVideosByCategoryAndType(ctx context.Context, contentType, category string, videos []response.VideoResponse, ttl time.Duration) error

	// DeleteVideosByCategoryAndType удаляет видео из кэша Redis
	DeleteVideosByCategoryAndType(ctx context.Context, contentType, category string) error
}
