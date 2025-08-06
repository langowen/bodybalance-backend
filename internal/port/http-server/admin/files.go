package admin

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin/dto"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
)

const (
	maxUploadSize      = 500 << 20 // 500 MB
	maxImageUploadSize = 20 << 20  // 20 MB
)

// @Summary Загрузить видеофайл
// @Description Загружает видеофайл на сервер (макс. 500MB)
// @Tags Admin Files
// @Accept multipart/form-data
// @Produce json
// @Param video formData file true "Видеофайл для загрузки"
// @Success 200 {object} dto.SuccessResponse "Видео успешно загружено"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security AdminAuth
// @Router /admin/files/video [post]
func (h *Handler) uploadVideoHandler(w http.ResponseWriter, r *http.Request) {
	const op = "admin.uploadVideoHandler"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	// Ограничиваем размер файла
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		logger.Error("File too large", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "File too large (max 500MB)")
		return
	}

	// Получаем файл из запроса
	file, header, err := r.FormFile("video")
	if err != nil {
		logger.Error("Failed to get file from request", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid file upload")
		return
	}
	defer file.Close()

	ctx := logging.ContextWithLogger(r.Context(), logger)

	err = h.service.UploadFile(ctx, file, header)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrFileTypeNotSupported):
			dto.RespondWithError(w, http.StatusBadRequest, "Invalid file type. Only video files are allowed")
			return
		case errors.Is(err, admin.ErrFailedToReadFile):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to read file")
			return
		case errors.Is(err, admin.ErrInvalidFileName):
			dto.RespondWithError(w, http.StatusBadRequest, "Имя файлов должны содержать только латинские буквы, цифры, дефисы и подчеркивания")
			return
		case errors.Is(err, admin.ErrFailedToSaveFile):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to save file")
			return
		}
	}

	res := dto.SuccessResponse{
		Message: fmt.Sprintf("Video %s uploaded successfully", header.Filename),
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Получить список видеофайлов
// @Description Возвращает список всех видеофайлов на сервере
// @Tags Admin Files
// @Produce json
// @Success 200 {array} dto.FileInfoResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security AdminAuth
// @Router /admin/files/video [get]
func (h *Handler) listVideoFilesHandler(w http.ResponseWriter, r *http.Request) {
	const op = "admin.listVideoFilesHandler"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	ctx := logging.ContextWithLogger(r.Context(), logger)

	files, err := h.service.ListVideoFiles(ctx)
	if err != nil {
		if errors.Is(err, admin.ErrFileNotFound) {
			dto.RespondWithError(w, http.StatusNotFound, "No video files found")
			return
		}
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get video files list")
		return
	}

	res := make([]dto.FileInfoResponse, len(files))
	for i, file := range files {
		res[i] = dto.FileInfoResponse{
			Name:    file.Name,
			Size:    file.Size,
			ModTime: file.ModTime,
		}
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Загрузить изображение
// @Description Загружает изображение на сервер (макс. 20MB)
// @Tags Admin Files
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Изображение для загрузки"
// @Success 200 {object} dto.SuccessResponse "Изображение успешно загружено"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security AdminAuth
// @Router /admin/files/img [post]
func (h *Handler) uploadImageHandler(w http.ResponseWriter, r *http.Request) {
	const op = "admin.uploadImageHandler"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	// Ограничиваем размер файла
	r.Body = http.MaxBytesReader(w, r.Body, maxImageUploadSize)
	if err := r.ParseMultipartForm(maxImageUploadSize); err != nil {
		logger.Error("Image too large", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Image too large (max 20MB)")
		return
	}

	// Получаем файл из запроса
	file, header, err := r.FormFile("image")
	if err != nil {
		logger.Error("Failed to get image from request", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid image upload")
		return
	}
	defer file.Close()

	ctx := logging.ContextWithLogger(r.Context(), logger)

	err = h.service.UploadImage(ctx, file, header)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrInvalidFileName):
			dto.RespondWithError(w, http.StatusBadRequest, "Имя файлов должны содержать только латинские буквы, цифры, дефисы и подчеркивания")
			return
		case errors.Is(err, admin.ErrFailedToReadFile):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to read image")
			return
		case errors.Is(err, admin.ErrFailedToSaveFile):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to save image")
			return
		case errors.Is(err, admin.ErrFileTypeNotSupported):
			dto.RespondWithError(w, http.StatusBadRequest, "Invalid image type. Only JPEG, PNG, GIF, SVG and WEBP are allowed")
			return
		}
	}

	res := dto.SuccessResponse{
		Message: fmt.Sprintf("Image %s uploaded successfully", header.Filename),
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Получить список изображений
// @Description Возвращает список всех изображений на сервере
// @Tags Admin Files
// @Produce json
// @Success 200 {array} dto.FileInfoResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security AdminAuth
// @Router /admin/files/img [get]
func (h *Handler) listImageFilesHandler(w http.ResponseWriter, r *http.Request) {
	const op = "admin.listImageFilesHandler"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	ctx := logging.ContextWithLogger(r.Context(), logger)

	files, err := h.service.ListImageFiles(ctx)
	if err != nil {
		if errors.Is(err, admin.ErrFileNotFound) {
			dto.RespondWithError(w, http.StatusNotFound, "No image files found")
			return
		}
		logger.Error("Failed to get images list", sl.Err(err))
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get images list")
		return
	}

	res := make([]dto.FileInfoResponse, len(files))
	for i, file := range files {
		res[i] = dto.FileInfoResponse{
			Name:    file.Name,
			Size:    file.Size,
			ModTime: file.ModTime,
		}
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}
