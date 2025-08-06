package admin

import "errors"

var (
	ErrEmptyUsername     = errors.New("username cannot be empty")
	ErrEmptyPassword     = errors.New("password cannot be empty")
	ErrUserNotAdmin      = errors.New("user is not an admin")
	ErrUserNotFound      = errors.New("user not found")
	ErrNotFoundPassword  = errors.New("password not found")
	ErrFailedSaveUser    = errors.New("failed to save user")
	ErrFailedGetUser     = errors.New("failed to get user")
	ErrFailedDeleteUser  = errors.New("failed to delete user")
	ErrUserInvalidID     = errors.New("invalid user ID")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type Users struct {
	ID          int64
	Username    string
	ContentType ContentType
	IsAdmin     bool
	Password    string
	DateCreated string
}
