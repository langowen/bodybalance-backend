package admin

import "errors"

type ContentType struct {
	ID         int64
	Name       string
	DataSource string
}

var (
	ErrEmptyTypeID = errors.New("type ID cannot be empty")
	ErrTypeInvalid = errors.New("invalid content type ID")
)
