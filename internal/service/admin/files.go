package admin

import (
	"context"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

const (
	videoMIMETypes = "video/mp4,video/quicktime,video/webm,video/ogg"
	imageMIMETypes = "image/jpeg,image/png,image/gif,image/webp,image/svg+xml"
)

func (s *ServiceAdmin) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) error {
	const op = "service.UploadFile"

	if !validFilePattern.MatchString(header.Filename) {
		logging.L(ctx).Warn("invalid file format in URL", "url", header.Filename, "op", op)
		return admin.ErrInvalidFileName
	}

	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		logging.L(ctx).Error("Failed to read file header", sl.Err(err), "op", op)
		return admin.ErrFailedToReadFile
	}

	mimeType := mimetype.Detect(buff)

	if !strings.Contains(videoMIMETypes, mimeType.String()) {
		logging.L(ctx).Error("Invalid image type", "content_type", mimeType.String(), "op", op)
		return admin.ErrFileTypeNotSupported
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logging.L(ctx).Error("Failed to reset file position", sl.Err(err), "op", op)
		return admin.ErrFailedToReadFile
	}

	if err := saveFile(header.Filename, file, s.cfg.Media.VideoPatch); err != nil {
		logging.L(ctx).Error("Failed to save image", sl.Err(err), "op", op)
		return admin.ErrFailedToSaveFile
	}

	return nil
}

func (s *ServiceAdmin) ListVideoFiles(ctx context.Context) ([]admin.File, error) {
	const op = "service.ListVideoFiles"

	files, err := getFilesList(ctx, s.cfg.Media.VideoPatch)
	if err != nil {
		logging.L(ctx).Error("Failed to read video directory", sl.Err(err), "op", op)
		return nil, err
	}

	if len(files) == 0 {
		logging.L(ctx).Warn("No video files found", "op", op)
		return nil, admin.ErrFileNotFound
	}

	return files, nil
}

func (s *ServiceAdmin) UploadImage(ctx context.Context, file multipart.File, header *multipart.FileHeader) error {
	const op = "service.UploadImage"

	if !validFilePattern.MatchString(header.Filename) {
		logging.L(ctx).Warn("invalid file format in URL", "url", header.Filename, "op", op)
		return admin.ErrInvalidFileName
	}

	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		logging.L(ctx).Error("Failed to read image header", sl.Err(err), "op", op)
		return admin.ErrFailedToReadFile
	}

	mimeType := mimetype.Detect(buff)

	if !strings.Contains(imageMIMETypes, mimeType.String()) {
		logging.L(ctx).Error("Invalid image type", "content_type", mimeType.String(), "op", op)
		return admin.ErrFileTypeNotSupported
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logging.L(ctx).Error("Failed to reset file position", sl.Err(err), "op", op)
		return admin.ErrFailedToReadFile
	}

	if err := saveFile(header.Filename, file, s.cfg.Media.ImagesPatch); err != nil {
		logging.L(ctx).Error("Failed to save image", sl.Err(err), "op", op)
		return admin.ErrFailedToSaveFile
	}

	return nil
}

func (s *ServiceAdmin) ListImageFiles(ctx context.Context) ([]admin.File, error) {
	const op = "service.ListImageFiles"

	files, err := getFilesList(ctx, s.cfg.Media.ImagesPatch)
	if err != nil {
		logging.L(ctx).Error("Failed to read image directory", sl.Err(err), "op", op)
		return nil, err
	}

	if len(files) == 0 {
		logging.L(ctx).Warn("No image files found", "op", op)
		return nil, admin.ErrFileNotFound
	}

	return files, nil
}

// saveFile сохраняет файлы
func saveFile(filename string, file multipart.File, patch string) error {
	if err := os.MkdirAll(patch, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	dst, err := os.Create(filepath.Join(patch, filename))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// getFilesList возвращает список файлов в указанной директории
func getFilesList(ctx context.Context, patch string) ([]admin.File, error) {
	files, err := os.ReadDir(patch)
	if err != nil {
		return nil, err
	}

	var result []admin.File
	for _, file := range files {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if file.IsDir() {
				continue
			}

			info, err := file.Info()
			if err != nil {
				continue
			}

			result = append(result, admin.File{
				Name:    file.Name(),
				Size:    info.Size(),
				ModTime: info.ModTime(),
			})
		}
	}

	return result, nil
}
