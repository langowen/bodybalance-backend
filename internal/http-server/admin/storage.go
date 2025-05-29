package admin

import (
	"context"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/langowen/bodybalance-backend/internal/storage/postgres/admin"
)

// AdmStorage определяет методы, которые нужны admin для работы с хранилищем.
type AdmStorage interface {
	AddVideo(ctx context.Context, video admResponse.VideoRequest) (int, error)
	GetVideo(ctx context.Context, id int64) (admResponse.VideoResponse, error)
	GetVideos(ctx context.Context) ([]admResponse.VideoResponse, error)
	UpdateVideo(ctx context.Context, id int64, video admResponse.VideoRequest) error
	DeleteVideo(ctx context.Context, id int64) error
	GetAdminUser(ctx context.Context, login, passwordHash string) (*admin.AdmUser, error)

	AddType(ctx context.Context, req admResponse.TypeRequest) (admResponse.TypeResponse, error)
	GetType(ctx context.Context, id int64) (admResponse.TypeResponse, error)
	GetTypes(ctx context.Context) ([]admResponse.TypeResponse, error)
	UpdateType(ctx context.Context, id int64, req admResponse.TypeRequest) error
	DeleteType(ctx context.Context, id int64) error
}
