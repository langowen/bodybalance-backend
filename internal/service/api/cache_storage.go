package api

import (
	"context"

	"github.com/langowen/bodybalance-backend/internal/entities/api"
)

type CacheStorageApi interface {
	GetCategories(ctx context.Context, typeID int64) ([]api.Category, error)
	SetCategories(ctx context.Context, typeID int64, categories []api.Category) error
	GetAccount(ctx context.Context, account *api.Account) (*api.Account, error)
	SetAccount(ctx context.Context, account *api.Account) error
	GetVideo(ctx context.Context, videoID int64) (*api.Video, error)
	SetVideo(ctx context.Context, videoID int64, video *api.Video) error
	GetVideosByCategoryAndType(ctx context.Context, typeID, catID int64) ([]api.Video, error)
	SetVideosByCategoryAndType(ctx context.Context, typeID, catID int64, videos []api.Video) error
	HealthCheck(ctx context.Context) error
}
