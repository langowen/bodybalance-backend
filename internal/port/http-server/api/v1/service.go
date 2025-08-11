package v1

import (
	"context"

	"github.com/langowen/bodybalance-backend/internal/entities/api"
)

type Service interface {
	GetTypeByAccount(ctx context.Context, username string) (*api.Account, error)
	GetCategoriesByType(ctx context.Context, contentType string) ([]api.Category, error)
	GetVideo(ctx context.Context, videoStr string) (*api.Video, error)
	GetVideosByCategoryAndType(ctx context.Context, contentType, category string) ([]api.Video, error)
	Feedback(ctx context.Context, feedback *api.Feedback) error
	HealthCheck(ctx context.Context) (*api.HealthCheck, error)
}
