package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	mwMetrics "github.com/langowen/bodybalance-backend/internal/http-server/middleware/metrics"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/langowen/bodybalance-backend/internal/storage"
	"github.com/theartofdevel/logging"
	"net/http"
	"strconv"
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

	videoStr := r.URL.Query().Get("video_id")

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
		"video_id", videoStr,
	)

	if videoStr == "" {
		logger.Error("Video id is empty")
		response.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Video id is empty")
		return
	}

	videoID, err := strconv.ParseInt(videoStr, 10, 64)
	if err != nil {
		logger.Error("Invalid video ID", sl.Err(err))
		response.RespondWithError(w, http.StatusBadRequest, "Bad Request", "invalid video ID",
			fmt.Sprintf("Video ID '%s' is not a valid number", videoStr))
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	var video *response.VideoResponse

	if h.cfg.Redis.Enable {
		video, err = h.redis.GetVideo(ctx, videoID)
		if err != nil {
			logger.Warn("failed to get video from redis", sl.Err(err))
		} else if video != nil {
			mwMetrics.RecordDataSource(r, mwMetrics.SourceRedis)

			logger.Debug("video fetched from redis cache")
			response.RespondWithJSON(w, http.StatusOK, video)
			return
		}
	}

	video, err = h.storage.GetVideo(ctx, videoID)
	if err != nil {
		if errors.Is(err, storage.ErrVideoNotFound) {
			response.RespondWithError(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Video with id %d not found", videoID))
			return
		}

		logger.Error("failed to get video", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Server Error", "Failed to retrieve video")
		return
	}

	mwMetrics.RecordDataSource(r, mwMetrics.SourceSQL)

	if h.cfg.Redis.Enable && video != nil {
		go func(ctx context.Context, videoID int64, video *response.VideoResponse) {
			if err := h.redis.SetVideo(ctx, videoID, video, h.cfg.Redis.CacheTTL); err != nil {
				logger.Warn("failed to cache video in redis", sl.Err(err))
			}
		}(ctx, videoID, video)
	}

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

	typeID, err := strconv.ParseInt(contentType, 10, 64)
	if err != nil {
		logger.Error("Invalid type ID", sl.Err(err))
		response.RespondWithError(w, http.StatusBadRequest, "Bad Request", "invalid type ID",
			fmt.Sprintf("Video ID '%s' is not a valid number", contentType))
		return
	}

	catID, err := strconv.ParseInt(categoryName, 10, 64)
	if err != nil {
		logger.Error("Invalid category ID", sl.Err(err))
		response.RespondWithError(w, http.StatusBadRequest, "Bad Request", "invalid category ID",
			fmt.Sprintf("Video ID '%s' is not a valid number", categoryName))
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	typeErr := h.storage.CheckType(ctx, typeID)
	if typeErr != nil {
		if errors.Is(typeErr, storage.ErrContentTypeNotFound) {
			logger.Warn("content type not found", sl.Err(typeErr))
			response.RespondWithError(w, http.StatusNotFound, "Not Found", "Content type not found", fmt.Sprintf("Content type '%d' does not exist", typeID))
			return
		}
		logger.Error("Failed to check content type", sl.Err(typeErr))
		response.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error", typeErr.Error())
		return
	}

	catErr := h.storage.CheckCategory(ctx, catID)
	if catErr != nil {
		if errors.Is(catErr, storage.ErrNoCategoriesFound) {
			logger.Warn("category not found", sl.Err(catErr))
			response.RespondWithError(w, http.StatusNotFound, "Not Found", "Category not found", fmt.Sprintf("Category '%d' does not exist", catID))
			return
		}
		logger.Error("Failed to check category", sl.Err(catErr))
		response.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error", catErr.Error())
		return
	}

	if h.cfg.Redis.Enable == true {
		cachedVideos, err := h.redis.GetVideosByCategoryAndType(ctx, typeID, catID)
		if err != nil {
			logger.Warn("redis get error", sl.Err(err))
		}

		if cachedVideos != nil {
			logger.Debug("send from cache")

			mwMetrics.RecordDataSource(r, mwMetrics.SourceRedis)

			response.RespondWithJSON(w, http.StatusOK, cachedVideos)
			return
		}
	}

	videos, err := h.storage.GetVideosByCategoryAndType(ctx, typeID, catID)
	switch {
	case errors.Is(err, storage.ErrContentTypeNotFound):
		logger.Warn("content type not found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Not Found", "Content type not found",
			fmt.Sprintf("Content type '%d' does not exist", typeID))
		return

	case errors.Is(err, storage.ErrNoCategoriesFound):
		logger.Warn("no categories found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Not Found", "Category not found",
			fmt.Sprintf("Category '%d' does not exist", catID))
		return

	case errors.Is(err, storage.ErrVideoNotFound):
		logger.Warn("video not found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Not Found", "Video not found",
			fmt.Sprintf("no videos found for content type '%d' and category '%d'",
				typeID, catID))
		return

	case err != nil:
		logger.Error("Failed to get videos", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	mwMetrics.RecordDataSource(r, mwMetrics.SourceSQL)

	if h.cfg.Redis.Enable == true {
		go func(ctx context.Context, typeID, catID int64, videos []response.VideoResponse) {
			if err := h.redis.SetVideosByCategoryAndType(ctx, typeID, catID, videos, h.cfg.Redis.CacheTTL); err != nil {
				logger.Warn("failed to set videos cache", sl.Err(err))
			}
		}(ctx, typeID, catID, videos)
	}

	response.RespondWithJSON(w, http.StatusOK, videos)
}
