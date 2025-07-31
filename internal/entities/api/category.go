package api

import "errors"

type Category struct {
	ID     int64
	Name   string
	ImgURL string
}

var (
	ErrEmptyCategoryID = errors.New("category ID cannot be empty")
	ErrCategoryInvalid = errors.New("invalid category ID")
)
