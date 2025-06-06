package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/langowen/bodybalance-backend/internal/storage"
	"github.com/theartofdevel/logging"
	"net/http"
	"time"
)

// @Summary Get video by ID
// @Description Returns video details by its ID
// @Tags API v1
// @Produce json
// @Param video_id query int true "Video ID"
// @Success 200 {object} response.VideoResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /video [get]
func (h *Handler) getVideo(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.getVideo"
	const cacheTTL = time.Hour * 24 // Время жизни кэша - 24 часа

	videoID := r.URL.Query().Get("video_id")

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
		"video_id", videoID,
	)

	if videoID == "" {
		logger.Error("Video id is empty")
		response.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Video id is empty")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	// Пытаемся получить данные из кэша
	cachedVideo, err := h.redis.GetVideo(ctx, videoID)
	if err != nil {
		logger.Warn("redis get error", sl.Err(err))
	}

	// Если данные есть в кэше - возвращаем их
	if cachedVideo != nil {
		logger.Debug("serving from cache", "video_id", videoID)
		response.RespondWithJSON(w, http.StatusOK, cachedVideo)
		return
	}

	// Данных нет в кэше - запрашиваем из основного хранилища
	video, err := h.storage.GetVideo(ctx, videoID)
	switch {
	case errors.Is(err, storage.ErrVideoNotFound):
		logger.Warn("video not found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Not Found", "Video not found",
			fmt.Sprintf("Video '%s' does not exist", videoID))
		return

	case err != nil:
		logger.Error("Failed to get video", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	// Сохраняем данные в кэш (в фоне, не блокируя ответ)
	go func() {
		ctx := context.Background() // используем новый контекст для фоновой операции
		if err := h.redis.SetVideo(ctx, videoID, &video, cacheTTL); err != nil {
			logger.Warn("failed to set video cache", sl.Err(err))
		}
	}()

	response.RespondWithJSON(w, http.StatusOK, video)
}

// @Summary Get videos by category and type
// @Description Returns videos filtered by type and category, ordered by name
// @Tags API v1
// @Produce json
// @Param type query int true "Type ID"
// @Param category query int true "Category ID"
// @Success 200 {array} response.VideoResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /video_categories [get]
func (h *Handler) getVideosByCategoryAndType(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.getVideosByCategoryAndType"
	const cacheTTL = time.Hour * 24

	contentType := r.URL.Query().Get("type")
	categoryName := r.URL.Query().Get("category")

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
		"type", contentType,
		"category", categoryName,
	)

	if categoryName == "" {
		logger.Error("Category is empty")
		response.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Category is empty")
		return
	}

	if contentType == "" {
		logger.Error("Content type is empty")
		response.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Content type is empty")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	// Пытаемся получить данные из кэша
	cachedVideos, err := h.redis.GetVideosByCategoryAndType(ctx, contentType, categoryName)
	if err != nil {
		logger.Warn("redis get error", sl.Err(err))
	}

	// Если данные есть в кэше - возвращаем их
	if cachedVideos != nil {
		logger.Debug("serving from cache",
			"content_type", contentType,
			"category", categoryName)
		response.RespondWithJSON(w, http.StatusOK, cachedVideos)
		return
	}

	// Данных нет в кэше - запрашиваем из основного хранилища
	videos, err := h.storage.GetVideosByCategoryAndType(ctx, contentType, categoryName)
	switch {
	case errors.Is(err, storage.ErrContentTypeNotFound):
		logger.Warn("content type not found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Not Found", "Content type not found",
			fmt.Sprintf("Content type '%s' does not exist", contentType))
		return

	case errors.Is(err, storage.ErrNoCategoriesFound):
		logger.Warn("no categories found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Not Found", "Category not found",
			fmt.Sprintf("Category '%s' does not exist", categoryName))
		return

	case errors.Is(err, storage.ErrVideoNotFound):
		logger.Warn("video not found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Not Found", "Video not found",
			fmt.Sprintf("no videos found for content type '%s' and category '%s'",
				contentType, categoryName))
		return

	case err != nil:
		logger.Error("Failed to get videos", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	// Сохраняем данные в кэш
	go func() {
		ctx := context.Background()
		if err := h.redis.SetVideosByCategoryAndType(ctx, contentType, categoryName, videos, cacheTTL); err != nil {
			logger.Warn("failed to set videos cache",
				sl.Err(err),
				"content_type", contentType,
				"category", categoryName)
		}
	}()

	response.RespondWithJSON(w, http.StatusOK, videos)
}
