package admin

import "errors"

var (
	ErrEmptyUsername = errors.New("username cannot be empty")
	ErrEmptyPassword = errors.New("password cannot be empty")
	ErrUserNotAdmin  = errors.New("user is not an admin")
	ErrUserNotFound  = errors.New("user not found")
)

type Users struct {
	Username    string
	ContentType ContentType
	IsAdmin     bool
	Password    string
}
