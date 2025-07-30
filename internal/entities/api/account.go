package api

import "errors"

var (
	ErrEmptyUsername      = errors.New("username cannot be empty")
	ErrStorageServerError = errors.New("storage server error")
)

type Account struct {
	Username    string
	ContentType ContentType
}
