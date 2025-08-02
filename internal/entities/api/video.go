package api

import "errors"

type Video struct {
	ID          int64
	URL         string
	Name        string
	Description string
	Category    Category
	ImgURL      string
	DataSource  string
}

var (
	ErrEmptyVideoID   = errors.New("video ID cannot be empty")
	ErrInvalidVideoID = errors.New("invalid video ID")
)
