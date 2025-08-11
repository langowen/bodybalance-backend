package api

import (
	"context"

	"github.com/langowen/bodybalance-backend/internal/entities/api"
)

// SqlStorageApi определяет методы, которые нужны API v1 для работы с хранилищем.
type SqlStorageApi interface {
	GetVideosByCategoryAndType(ctx context.Context, TypeID, CatID int64) ([]api.Video, error)
	GetCategories(ctx context.Context, TypeID int64) ([]api.Category, error)
	CheckAccount(ctx context.Context, account *api.Account) (*api.Account, error)
	GetVideo(ctx context.Context, videoID int64) (*api.Video, error)
	Feedback(ctx context.Context, feedback *api.Feedback) error
	HealthCheck(ctx context.Context) error
}
