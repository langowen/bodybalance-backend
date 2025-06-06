package v1

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/langowen/bodybalance-backend/internal/storage"
	"github.com/theartofdevel/logging"
	"net/http"
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

	videoID := r.URL.Query().Get("video_id")

	// Создаем логгер с дополнительными полями
	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
		"video_id", videoID,
	)

	if videoID == "" {
		logger.Error("Video id is empty")
		response.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Video id is empty")
	}

	// Создаем новый контекст с логгером
	ctx := logging.ContextWithLogger(r.Context(), logger)

	videos, err := h.storage.GetVideo(ctx, videoID)
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

	response.RespondWithJSON(w, http.StatusOK, videos)
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

	contentType := r.URL.Query().Get("type")
	categoryName := r.URL.Query().Get("category")

	// Создаем логгер с дополнительными полями
	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
		"type", contentType,
		"category", categoryName,
	)

	if categoryName == "" {
		logger.Error("Category is empty")
		response.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Category is empty")
	}

	if contentType == "" {
		logger.Error("Content type is empty")
		response.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Content type is empty")
	}

	// Создаем новый контекст с логгером
	ctx := logging.ContextWithLogger(r.Context(), logger)

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

	response.RespondWithJSON(w, http.StatusOK, videos)
}
