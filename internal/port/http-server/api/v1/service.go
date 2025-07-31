package v1

import (
	"context"
	"github.com/langowen/bodybalance-backend/internal/entities/api"
)

type Service interface {
	GetTypeByAccount(ctx context.Context, username string) (*api.Account, string, error)
	GetCategoriesByType(ctx context.Context, contentType string) ([]api.Category, string, error)
	GetVideo(ctx context.Context, videoStr string) (*api.Video, string, error)
	GetVideosByCategoryAndType(ctx context.Context, contentType, category string) ([]api.Video, string, error)
	Feedback(ctx context.Context, feedback *api.Feedback) error
}
