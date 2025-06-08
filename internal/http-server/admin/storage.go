package admin

import (
	"context"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/storage/postgres/admin"
	"time"
)

// AdmStorage определяет методы, которые нужны admin для работы с хранилищем.
type AdmStorage interface {
	AddVideo(ctx context.Context, video admResponse.VideoRequest) (int64, error)
	GetVideo(ctx context.Context, id int64) (admResponse.VideoResponse, error)
	GetVideos(ctx context.Context) ([]admResponse.VideoResponse, error)
	UpdateVideo(ctx context.Context, id int64, video admResponse.VideoRequest) error
	DeleteVideo(ctx context.Context, id int64) error
	DeleteVideoCategories(ctx context.Context, videoID int64) error
	GetVideoCategories(ctx context.Context, videoID int64) ([]admResponse.CategoryResponse, error)
	AddVideoCategories(ctx context.Context, videoID int64, categoryIDs []int64) error

	GetAdminUser(ctx context.Context, login, passwordHash string) (*admin.AdmUser, error)

	AddType(ctx context.Context, req admResponse.TypeRequest) (admResponse.TypeResponse, error)
	GetType(ctx context.Context, id int64) (admResponse.TypeResponse, error)
	GetTypes(ctx context.Context) ([]admResponse.TypeResponse, error)
	UpdateType(ctx context.Context, id int64, req admResponse.TypeRequest) error
	DeleteType(ctx context.Context, id int64) error

	AddUser(ctx context.Context, req admResponse.UserRequest) (admResponse.UserResponse, error)
	GetUser(ctx context.Context, id int64) (admResponse.UserResponse, error)
	GetUsers(ctx context.Context) ([]admResponse.UserResponse, error)
	UpdateUser(ctx context.Context, id int64, req admResponse.UserRequest) error
	DeleteUser(ctx context.Context, id int64) error

	AddCategory(ctx context.Context, req admResponse.CategoryRequest) (admResponse.CategoryResponse, error)
	GetCategory(ctx context.Context, id int64) (admResponse.CategoryResponse, error)
	GetCategories(ctx context.Context) ([]admResponse.CategoryResponse, error)
	UpdateCategory(ctx context.Context, id int64, req admResponse.CategoryRequest) error
	DeleteCategory(ctx context.Context, id int64) error
}

// RedisStorage интерфейс для работы с Redis хранилищем
type RedisStorage interface {
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
