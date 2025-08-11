package api

import "errors"

var (
	ErrEmptyUsername      = errors.New("username cannot be empty")
	ErrStorageServerError = errors.New("storage server error")
	ErrRedisError         = errors.New("redis server error")
)

type Account struct {
	Username    string
	ContentType ContentType
	DataSource  string
}
