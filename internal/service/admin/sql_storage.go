package admin

import (
	"context"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
)

// AdmStorage определяет методы, которые нужны admin для работы с хранилищем.
type AdmStorage interface {
	AddVideo(ctx context.Context, video *admin.Video) (int64, error)
	GetVideo(ctx context.Context, id int64) (*admin.Video, error)
	GetVideos(ctx context.Context) ([]admin.Video, error)
	UpdateVideo(ctx context.Context, video *admin.Video) error
	DeleteVideo(ctx context.Context, id int64) error

	GetAdminUser(ctx context.Context, login, passwordHash string) (*admin.Users, error)

	AddType(ctx context.Context, req *admin.ContentType) (*admin.ContentType, error)
	GetType(ctx context.Context, id int64) (*admin.ContentType, error)
	GetTypes(ctx context.Context) ([]admin.ContentType, error)
	UpdateType(ctx context.Context, req *admin.ContentType) error
	DeleteType(ctx context.Context, id int64) error

	AddUser(ctx context.Context, req *admin.Users) (*admin.Users, error)
	GetUser(ctx context.Context, id int64) (*admin.Users, error)
	GetUsers(ctx context.Context) ([]admin.Users, error)
	UpdateUser(ctx context.Context, req *admin.Users) error
	DeleteUser(ctx context.Context, id int64) error

	AddCategory(ctx context.Context, req *admin.Category) (*admin.Category, error)
	GetCategory(ctx context.Context, id int64) (*admin.Category, error)
	GetCategories(ctx context.Context) ([]admin.Category, error)
	UpdateCategory(ctx context.Context, id int64, req *admin.Category) error
	DeleteCategory(ctx context.Context, id int64) error
}
