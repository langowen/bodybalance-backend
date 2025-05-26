package v1

import (
	"context"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
)

// ApiStorage определяет методы, которые нужны API v1 для работы с хранилищем.
type ApiStorage interface {
	GetVideosByCategoryAndType(ctx context.Context, contentType, categoryName string) ([]response.VideoResponse, error)
	GetCategories(ctx context.Context, contentType string) ([]response.CategoryResponse, error)
	CheckAccount(ctx context.Context, username string) (response.AccountResponse, error)
	GetVideo(ctx context.Context, videoID string) (response.VideoResponse, error)
}
