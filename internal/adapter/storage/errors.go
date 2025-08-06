package storage

import (
	"errors"
)

var (
	ErrContentTypeNotFound = errors.New("content type not found")
	ErrNoCategoriesFound   = errors.New("no categories found")
	ErrVideoNotFound       = errors.New("no video found")
	ErrAccountNotFound     = errors.New("account not found")
)
