package storage

import "context"

type ApiStorage interface {
	GetVideosByCategoryAndType(ctx context.Context, contentType, categoryName string) ([]Video, error)
	GetCategoriesWithVideos(ctx context.Context, contentType string) ([]CategoryWithVideos, error)
	CheckAccount(ctx context.Context, username string) (bool, error)
}

// Video represents video item response structure
type Video struct {
	ID          float64 `json:"id"`
	URL         string  `json:"url"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
}

// CategoryWithVideos представляет категорию с ID из БД
type CategoryWithVideos struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Videos []Video `json:"videos,omitempty"`
}
