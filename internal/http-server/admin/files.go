package admin

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxUploadSize      = 500 << 20 // 500 MB
	videoMIMETypes     = "video/mp4,video/webm,video/ogg"
	maxImageUploadSize = 20 << 20 // 20 MB
	imageMIMETypes     = "image/jpeg,image/png,image/gif,image/webp"
)

// @Summary Загрузить видеофайл
// @Description Загружает видеофайл на сервер (макс. 500MB)
// @Tags Admin Files
// @Accept multipart/form-data
// @Produce json
// @Param video formData file true "Видеофайл для загрузки"
// @Success 200 {object} object{message=string} "Файл успешно загружен"
// @Failure 400 {object} admResponse.ErrorResponse
// @Failure 500 {object} admResponse.ErrorResponse
// @Security AdminAuth
// @Router /admin/files/video [post]
func (h *Handler) uploadVideoHandler(w http.ResponseWriter, r *http.Request) {
	const op = "admin.uploadVideoHandler"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	// Ограничиваем размер файла
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		logger.Error("File too large", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "File too large (max 500MB)")
		return
	}

	// Получаем файл из запроса
	file, header, err := r.FormFile("video")
	if err != nil {
		logger.Error("Failed to get file from request", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid file upload")
		return
	}
	defer file.Close()

	// Проверяем MIME-тип
	buff := make([]byte, 512)
	if _, err = file.Read(buff); err != nil {
		logger.Error("Failed to read file header", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to verify file type")
		return
	}

	if !strings.Contains(videoMIMETypes, http.DetectContentType(buff)) {
		logger.Error("Invalid file type", "content_type", http.DetectContentType(buff))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid file type. Only video files are allowed")
		return
	}

	// Сбрасываем позицию чтения файла
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		logger.Error("Failed to reset file position", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to process file")
		return
	}

	// Сохраняем файл
	if err := h.saveVideoFile(header.Filename, file); err != nil {
		logger.Error("Failed to save file", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("File %s uploaded successfully", header.Filename),
	})
}

// @Summary Получить список видеофайлов
// @Description Возвращает список всех видеофайлов на сервере
// @Tags Admin Files
// @Produce json
// @Success 200 {array} admResponse.FileInfo
// @Failure 500 {object} admResponse.ErrorResponse
// @Security AdminAuth
// @Router /admin/files/video [get]
func (h *Handler) listVideoFilesHandler(w http.ResponseWriter, r *http.Request) {
	const op = "admin.listVideoFilesHandler"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	files, err := h.getVideoFilesList()
	if err != nil {
		logger.Error("Failed to get files list", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to get files list")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, files)
}

// saveVideoFile сохраняет видеофайл
func (h *Handler) saveVideoFile(filename string, file multipart.File) error {
	// Проверяем имя файла
	if !isValidFilename(filename) {
		return errors.New("invalid filename")
	}

	// Создаем папку если не существует
	if err := os.MkdirAll(h.cfg.Media.VideoPatch, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Создаем файл на сервере
	dst, err := os.Create(filepath.Join(h.cfg.Media.VideoPatch, filename))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Копируем содержимое
	if _, err := io.Copy(dst, file); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// getVideoFilesList возвращает список видеофайлов
func (h *Handler) getVideoFilesList() ([]admResponse.FileInfo, error) {
	files, err := os.ReadDir(h.cfg.Media.VideoPatch)
	if err != nil {
		return nil, err
	}

	var result []admResponse.FileInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// Проверяем расширение файла
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if !isVideoExtension(ext) {
			continue
		}

		result = append(result, admResponse.FileInfo{
			Name:    file.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	return result, nil
}

// @Summary Загрузить изображение
// @Description Загружает изображение на сервер (макс. 20MB)
// @Tags Admin Files
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Изображение для загрузки"
// @Success 200 {object} object{message=string} "Изображение успешно загружено"
// @Failure 400 {object} admResponse.ErrorResponse
// @Failure 500 {object} admResponse.ErrorResponse
// @Security AdminAuth
// @Router /admin/files/img [post]
func (h *Handler) uploadImageHandler(w http.ResponseWriter, r *http.Request) {
	const op = "admin.uploadImageHandler"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	// Ограничиваем размер файла
	r.Body = http.MaxBytesReader(w, r.Body, maxImageUploadSize)
	if err := r.ParseMultipartForm(maxImageUploadSize); err != nil {
		logger.Error("Image too large", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Image too large (max 20MB)")
		return
	}

	// Получаем файл из запроса
	file, header, err := r.FormFile("image")
	if err != nil {
		logger.Error("Failed to get image from request", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid image upload")
		return
	}
	defer file.Close()

	// Проверяем MIME-тип
	buff := make([]byte, 512)
	if _, err = file.Read(buff); err != nil {
		logger.Error("Failed to read image header", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to verify image type")
		return
	}

	if !strings.Contains(imageMIMETypes, http.DetectContentType(buff)) {
		logger.Error("Invalid image type", "content_type", http.DetectContentType(buff))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid image type. Only JPEG, PNG, GIF and WEBP are allowed")
		return
	}

	// Сбрасываем позицию чтения файла
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		logger.Error("Failed to reset file position", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to process image")
		return
	}

	// Сохраняем файл
	if err := h.saveImageFile(header.Filename, file); err != nil {
		logger.Error("Failed to save image", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to save image")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Image %s uploaded successfully", header.Filename),
	})
}

// @Summary Получить список изображений
// @Description Возвращает список всех изображений на сервере
// @Tags Admin Files
// @Produce json
// @Success 200 {array} admResponse.FileInfo
// @Failure 500 {object} admResponse.ErrorResponse
// @Security AdminAuth
// @Router /admin/files/img [get]
func (h *Handler) listImageFilesHandler(w http.ResponseWriter, r *http.Request) {
	const op = "admin.listImageFilesHandler"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	files, err := h.getImageFilesList()
	if err != nil {
		logger.Error("Failed to get images list", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to get images list")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, files)
}

// saveImageFile сохраняет изображение
func (h *Handler) saveImageFile(filename string, file multipart.File) error {
	// Проверяем имя файла
	if !isValidFilename(filename) {
		return errors.New("invalid filename")
	}

	// Создаем папку если не существует
	if err := os.MkdirAll(h.cfg.Media.ImagesPatch, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Создаем файл на сервере
	dst, err := os.Create(filepath.Join(h.cfg.Media.ImagesPatch, filename))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Копируем содержимое
	if _, err := io.Copy(dst, file); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// getImageFilesList возвращает список изображений
func (h *Handler) getImageFilesList() ([]admResponse.FileInfo, error) {
	files, err := os.ReadDir(h.cfg.Media.ImagesPatch)
	if err != nil {
		return nil, err
	}

	var result []admResponse.FileInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// Проверяем расширение файла
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if !isImageExtension(ext) {
			continue
		}

		result = append(result, admResponse.FileInfo{
			Name:    file.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	return result, nil
}

// Вспомогательные функции для изображений
func isImageExtension(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	}
	return false
}

// Вспомогательные функции для видео
func isValidFilename(filename string) bool {
	return !strings.ContainsAny(filename, "\\/:*?\"<>|")
}

func isVideoExtension(ext string) bool {
	switch ext {
	case ".mp4", ".webm", ".ogg", ".mov", ".avi":
		return true
	}
	return false
}
