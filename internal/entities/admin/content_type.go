package admin

import "errors"

type ContentType struct {
	ID          int64
	Name        string
	DateCreated string
}

var (
	ErrTypeInvalid    = errors.New("invalid content type ID")
	ErrTypeNotFound   = errors.New("content type not found")
	ErrTypeNameEmpty  = errors.New("content type name cannot be empty")
	ErrFailedSaveType = errors.New("failed to save content type")
	ErrFailedGetType  = errors.New("failed to get content type")
)
