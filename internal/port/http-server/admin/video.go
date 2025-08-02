package admin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin/dto"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// validFilePattern паттерны для проверки правильности названия файлов
var validFilePattern = regexp.MustCompile(`^[a-zA-Z0-9а-яА-ЯёЁ_\-.]+\.[a-zA-Z0-9]+$`)

// suspiciousPatterns паттерны для проверки нет ли лишних символов и ссылок в данных
var suspiciousPatterns = []string{"://", "//", "../", "./", "\\", "?", "&", "=", "%"}

// @Summary Добавить новое видео
// @Description Создает новую запись видео в системе
// @Tags Admin Videos
// @Accept json
// @Produce json
// @Param input body dto.VideoRequest true "Данные видео"
// @Success 201 {object} object{id=int64,message=string} "Видео успешно добавлено"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/video [post]
func (h *Handler) addVideo(w http.ResponseWriter, r *http.Request) {
	const op = "admin.addVideo"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	var req dto.VideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if !h.validVideo(&req, w, logger) {
		return
	}

	ctx := r.Context()
	videoID, err := h.storage.AddVideo(ctx, &req)
	if err != nil {
		logger.Error("failed to add video", sl.Err(err))
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to add video")
		return
	}

	if len(req.CategoryIDs) > 0 {
		err = h.storage.AddVideoCategories(ctx, videoID, req.CategoryIDs)
		if err != nil {
			logger.Error("failed to add video categories", sl.Err(err))
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to add video categories")
			return
		}
	}

	if h.cfg.Redis.Enable == true {
		go h.removeCache(op)
	}

	dto.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      videoID,
		"message": "Video added successfully",
	})
}

// @Summary Получить видео по ID
// @Description Возвращает информацию о конкретном видео
// @Tags Admin Videos
// @Produce json
// @Param id path int true "ID видео"
// @Success 200 {object} dto.VideoResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/video/{id} [get]
func (h *Handler) getVideo(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getVideo"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid video ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
		return
	}

	ctx := r.Context()
	video, err := h.storage.GetVideo(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("video not found", "video_id", id)
			dto.RespondWithError(w, http.StatusNotFound, "Video not found")
			return
		}
		logger.Error("failed to get video", sl.Err(err), "video_id", id)
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get video")
		return
	}

	categories, err := h.storage.GetVideoCategories(ctx, id)
	if err != nil {
		logger.Error("failed to get video categories", sl.Err(err), "video_id", id)
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get video categories")
		return
	}

	if len(categories) != 0 {
		video.Categories = categories
	} else {
		video.Categories = []dto.CategoryResponse{}
	}

	dto.RespondWithJSON(w, http.StatusOK, video)
}

// @Summary Получить список всех видео
// @Description Возвращает список всех видео в системе
// @Tags Admin Videos
// @Produce json
// @Success 200 {array} dto.VideoResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/video [get]
func (h *Handler) getVideos(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getVideos"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	ctx := r.Context()
	videos, err := h.storage.GetVideos(ctx)
	if err != nil {
		logger.Error("failed to get videos", sl.Err(err))
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get videos")
		return
	}

	for i := range videos {
		categories, err := h.storage.GetVideoCategories(ctx, videos[i].ID)
		if err != nil {
			logger.Error("failed to get video categories", sl.Err(err), "video_id", videos[i].ID)
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get video categories")
			return
		}

		if len(videos[i].Categories) == 0 {
			videos[i].Categories = make([]dto.CategoryResponse, 0, len(categories))
		}

		videos[i].Categories = append(videos[i].Categories, categories...)
	}

	dto.RespondWithJSON(w, http.StatusOK, videos)
}

// @Summary Обновить данные видео
// @Description Обновляет информацию о существующем видео
// @Tags Admin Videos
// @Accept json
// @Produce json
// @Param id path int true "ID видео"
// @Param input body dto.VideoRequest true "Новые данные видео"
// @Success 200 {object} object{id=int64,message=string} "Видео успешно обновлено"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/video/{id} [put]
func (h *Handler) updateVideo(w http.ResponseWriter, r *http.Request) {
	const op = "admin.updateVideo"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid video ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
		return
	}

	ctx := r.Context()

	var req dto.VideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if !h.validVideo(&req, w, logger) {
		return
	}

	err = h.storage.UpdateVideo(ctx, id, &req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("video not found", "video_id", id)
			dto.RespondWithError(w, http.StatusNotFound, "Video not found")
			return
		}
		logger.Error("failed to update video", sl.Err(err), "video_id", id)
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to update video")
		return
	}

	// Обновляем категории видео
	if len(req.CategoryIDs) > 0 {
		// Сначала удаляем все существующие связи
		err = h.storage.DeleteVideoCategories(ctx, id)
		if err != nil {
			logger.Error("failed to delete video categories", sl.Err(err), "video_id", id)
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to update video categories")
			return
		}

		// Затем добавляем новые
		err = h.storage.AddVideoCategories(ctx, id, req.CategoryIDs)
		if err != nil {
			logger.Error("failed to add video categories", sl.Err(err), "video_id", id)
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to update video categories")
			return
		}
	}

	if h.cfg.Redis.Enable == true {
		go h.removeCache(op)
	}

	dto.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Video updated successfully",
	})
}

// @Summary Удалить видео
// @Description Помечает видео как удаленное в системе
// @Tags Admin Videos
// @Produce json
// @Param id path int true "ID видео"
// @Success 200 {object} object{id=int64,message=string} "Видео успешно удалено"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/video/{id} [delete]
func (h *Handler) deleteVideo(w http.ResponseWriter, r *http.Request) {
	const op = "admin.deleteVideo"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid video ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
		return
	}

	ctx := r.Context()

	err = h.storage.DeleteVideo(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("video not found", "video_id", id)
			dto.RespondWithError(w, http.StatusNotFound, "Video not found")
			return
		}
		logger.Error("failed to delete video", sl.Err(err), "video_id", id)
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to delete video")
		return
	}

	if h.cfg.Redis.Enable == true {
		go h.removeCache(op)
	}

	dto.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Video deleted successfully",
	})
}

// validUser проверят входящие данные на валидность
func (h *Handler) validVideo(req *dto.VideoRequest, w http.ResponseWriter, logger *logging.Logger) bool {
	switch {
	case req.Name == "":
		logger.Warn("empty required Name")
		dto.RespondWithError(w, http.StatusBadRequest, "Введите название видео")
		return false
	case req.URL == "":
		logger.Warn("empty required video")
		dto.RespondWithError(w, http.StatusBadRequest, "Выберете видео")
		return false
	case req.ImgURL == "":
		logger.Warn("empty required img")
		dto.RespondWithError(w, http.StatusBadRequest, "Выберите превью для видео")
		return false
	case len(req.CategoryIDs) == 0:
		logger.Warn("empty required CategoryIDs")
		dto.RespondWithError(w, http.StatusBadRequest, "Выберите хотя бы одну категорию для видео")
		return false
	}

	if !validFilePattern.MatchString(req.URL) {
		logger.Warn("invalid file format in URL", "url", req.URL)
		dto.RespondWithError(w, http.StatusBadRequest, "Недопустимый формат имени видео-файла")
		return false
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(req.URL, pattern) {
			logger.Warn("suspicious pattern in URL", "url", req.URL, "pattern", pattern)
			dto.RespondWithError(w, http.StatusBadRequest, "Недопустимые символы в имени видео-файла")
			return false
		}
	}

	if !validFilePattern.MatchString(req.ImgURL) {
		logger.Warn("invalid file format in ImgURL", "imgurl", req.ImgURL)
		dto.RespondWithError(w, http.StatusBadRequest, "Недопустимый формат имени файла превью")
		return false
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(req.ImgURL, pattern) {
			logger.Warn("suspicious pattern in ImgURL", "imgurl", req.ImgURL, "pattern", pattern)
			dto.RespondWithError(w, http.StatusBadRequest, "Недопустимые символы в имени файла превью")
			return false
		}
	}

	return true
}
