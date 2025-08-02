package admin

import (
	"context"
	"github.com/langowen/bodybalance-backend/internal/adapter/storage/postgres/admin"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin/dto"
)

// AdmStorage определяет методы, которые нужны admin для работы с хранилищем.
type AdmStorage interface {
	AddVideo(ctx context.Context, video *dto.VideoRequest) (int64, error)
	GetVideo(ctx context.Context, id int64) (*dto.VideoResponse, error)
	GetVideos(ctx context.Context) ([]dto.VideoResponse, error)
	UpdateVideo(ctx context.Context, id int64, video *dto.VideoRequest) error
	DeleteVideo(ctx context.Context, id int64) error
	DeleteVideoCategories(ctx context.Context, videoID int64) error
	GetVideoCategories(ctx context.Context, videoID int64) ([]dto.CategoryResponse, error)
	AddVideoCategories(ctx context.Context, videoID int64, categoryIDs []int64) error

	GetAdminUser(ctx context.Context, login, passwordHash string) (*admin.AdmUser, error)

	AddType(ctx context.Context, req *dto.TypeRequest) (*dto.TypeResponse, error)
	GetType(ctx context.Context, id int64) (*dto.TypeResponse, error)
	GetTypes(ctx context.Context) ([]dto.TypeResponse, error)
	UpdateType(ctx context.Context, id int64, req *dto.TypeRequest) error
	DeleteType(ctx context.Context, id int64) error

	AddUser(ctx context.Context, req *dto.UserRequest) (*dto.UserResponse, error)
	GetUser(ctx context.Context, id int64) (*dto.UserResponse, error)
	GetUsers(ctx context.Context) ([]dto.UserResponse, error)
	UpdateUser(ctx context.Context, id int64, req *dto.UserRequest) error
	DeleteUser(ctx context.Context, id int64) error

	AddCategory(ctx context.Context, req *dto.CategoryRequest) (*dto.CategoryResponse, error)
	GetCategory(ctx context.Context, id int64) (*dto.CategoryResponse, error)
	GetCategories(ctx context.Context) ([]dto.CategoryResponse, error)
	UpdateCategory(ctx context.Context, id int64, req *dto.CategoryRequest) error
	DeleteCategory(ctx context.Context, id int64) error
}
