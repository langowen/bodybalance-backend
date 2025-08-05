package admin

import "errors"

var (
	ErrEmptyName         = errors.New("name cannot be empty")
	ErrEmptyImgURL       = errors.New("image URL cannot be empty")
	ErrEmptyTypeIDs      = errors.New("at least one content type ID must be selected")
	ErrInvalidImgFormat  = errors.New("invalid file format in image URL")
	ErrSuspiciousContent = errors.New("suspicious pattern in image URL")
	ErrCategoryNotFound  = errors.New("category not found")
)

type Category struct {
	ID          int64
	Name        string
	ImgURL      string
	ContentType []ContentType
	CreatedAt   string
}
