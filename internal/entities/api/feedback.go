package api

import "errors"

type Feedback struct {
	Name     string
	Email    string
	Telegram string
	Message  string
}

var (
	ErrEmptyTelegramOrEmail = errors.New("either telegram or email must be provided")
	ErrEmptyMessage         = errors.New("message cannot be empty")
	ErrInvalidEmail         = errors.New("invalid email format")
	ErrInvalidTelegram      = errors.New("invalid telegram format")
)
