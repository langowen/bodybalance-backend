package storage

import (
	"context"
	"errors"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
)

var (
	ErrContentTypeNotFound = errors.New("content type not found")
	ErrNoCategoriesFound   = errors.New("no categories found")
	ErrVideoNotFound       = errors.New("no video found")
)

type ApiStorage interface {
	GetVideosByCategoryAndType(ctx context.Context, contentType, categoryName string) ([]response.VideoResponse, error)
	GetCategories(ctx context.Context, contentType string) ([]response.CategoryResponse, error)
	CheckAccount(ctx context.Context, username string) (bool, error)
	GetVideo(ctx context.Context, videoID string) (response.VideoResponse, error)
}
