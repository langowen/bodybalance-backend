package admin

import "errors"

var (
	ErrVideoNotFound             = errors.New("video not found")
	ErrVideoInvalidID            = errors.New("invalid video ID")
	ErrVideoInvalidURL           = errors.New("invalid video URL")
	ErrVideoInvalidName          = errors.New("video name cannot be empty")
	ErrVideoInvalidImgURL        = errors.New("video image URL cannot be empty")
	ErrVideoInvalidCategory      = errors.New("at least one category must be selected")
	ErrVideoURLPattern           = errors.New("invalid file format in video URL")
	ErrVideoSuspiciousPattern    = errors.New("suspicious pattern in video URL")
	ErrVideoImgPattern           = errors.New("invalid file format in video image URL")
	ErrVideoImgSuspiciousPattern = errors.New("suspicious pattern in video image URL")
	ErrVideoSaveFailed           = errors.New("failed to save video")
	ErrFailedGetVideo            = errors.New("failed to get video")
	ErrVideoDeleteFailed         = errors.New("failed to delete video")
	ErrVideoUpdateFailed         = errors.New("failed to update video")
)

type Video struct {
	ID          int64
	URL         string
	Name        string
	Description string
	ImgURL      string
	Categories  []Category
	DateCreated string
}
