package v1

import (
	"context"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
)

// ApiStorage определяет методы, которые нужны API v1 для работы с хранилищем.
type ApiStorage interface {
	GetVideosByCategoryAndType(ctx context.Context, TypeID, CatID int64) (*[]response.VideoResponse, error)
	GetCategories(ctx context.Context, TypeID int64) (*[]response.CategoryResponse, error)
	CheckAccount(ctx context.Context, username string) (*response.AccountResponse, error)
	GetVideo(ctx context.Context, videoID int64) (*response.VideoResponse, error)
	// existence check for type and category
	CheckType(ctx context.Context, TypeID int64) error
	CheckCategory(ctx context.Context, CatID int64) error
}
