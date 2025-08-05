package admin

import (
	"context"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
	"mime/multipart"
)

type Service interface {
	// Video methods
	AddVideo(ctx context.Context, req *admin.Video) (int64, error)
	GetVideo(ctx context.Context, id int64) (*admin.Video, error)
	GetVideos(ctx context.Context) ([]admin.Video, error)
	UpdateVideo(ctx context.Context, req *admin.Video) error
	DeleteVideo(ctx context.Context, id int64) error
	// User methods
	AddUser(ctx context.Context, req *admin.Users) (*admin.Users, error)
	GetUser(ctx context.Context, id int64) (*admin.Users, error)
	GetUsers(ctx context.Context) ([]admin.Users, error)
	UpdateUser(ctx context.Context, req *admin.Users) error
	DeleteUser(ctx context.Context, id int64) error
	// Type methods
	AddType(ctx context.Context, req *admin.ContentType) (*admin.ContentType, error)
	GetType(ctx context.Context, id int64) (*admin.ContentType, error)
	GetTypes(ctx context.Context) ([]admin.ContentType, error)
	UpdateType(ctx context.Context, req *admin.ContentType) error
	DeleteType(ctx context.Context, id int64) error
	// Category methods
	AddCategory(ctx context.Context, req *admin.Category) (*admin.Category, error)
	GetCategory(ctx context.Context, id int64) (*admin.Category, error)
	GetCategories(ctx context.Context) ([]admin.Category, error)
	UpdateCategory(ctx context.Context, id int64, req *admin.Category) error
	DeleteCategory(ctx context.Context, id int64) error
	// Auth methods
	Signing(ctx context.Context, login, password string) (*admin.Users, error)
	// File methods
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) error
	ListVideoFiles(ctx context.Context) ([]admin.File, error)
	UploadImage(ctx context.Context, file multipart.File, header *multipart.FileHeader) error
	ListImageFiles(ctx context.Context) ([]admin.File, error)
}
