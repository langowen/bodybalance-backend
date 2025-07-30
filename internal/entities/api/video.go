package api

import "errors"

type Video struct {
	ID          int64
	URL         string
	Name        string
	Description string
	Category    Category
	ImgURL      string
}

var (
	ErrEmptyVideoID    = errors.New("video ID cannot be empty")
	ErrInvalidVideoID  = errors.New("invalid video ID")
	ErrEmptyCategoryID = errors.New("category ID cannot be empty")
	ErrCategoryInvalid = errors.New("invalid category ID")
)
