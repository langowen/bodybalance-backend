package admin

import (
	"errors"
	"time"
)

var (
	ErrFileTypeNotSupported = errors.New("file type not supported")
	ErrFailedToSaveFile     = errors.New("failed to save file")
	ErrFailedToReadFile     = errors.New("failed to read file")
	ErrInvalidFileName      = errors.New("invalid file name")
	ErrFileNotFound         = errors.New("file not found")
)

type File struct {
	Name    string
	Size    int64
	ModTime time.Time
}
