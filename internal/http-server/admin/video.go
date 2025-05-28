package admin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"net/http"
	"strconv"
)

// addVideo добавляет новое видео в БД
func (h *Handler) addVideo(w http.ResponseWriter, r *http.Request) {
	const op = "admin.addVideo"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	var req admResponse.VideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.URL == "" || req.Name == "" {
		logger.Error("empty required fields")
		admResponse.RespondWithError(w, http.StatusBadRequest, "URL and Name are required")
		return
	}

	ctx := r.Context()
	videoID, err := h.storage.AddVideo(ctx, req)
	if err != nil {
		logger.Error("failed to add video", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to add video")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      videoID,
		"message": "Video added successfully",
	})
}

// getVideo возвращает видео по его ID
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
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
		return
	}

	ctx := r.Context()
	video, err := h.storage.GetVideo(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("video not found", "video_id", id)
			admResponse.RespondWithError(w, http.StatusNotFound, "Video not found")
			return
		}
		logger.Error("failed to get video", sl.Err(err), "video_id", id)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to get video")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, video)
}

// getVideos возвращает все видео
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
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to get videos")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, videos)
}

// updateVideo обновляет данные видео
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
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
		return
	}

	var req admResponse.VideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	ctx := r.Context()
	err = h.storage.UpdateVideo(ctx, id, req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("video not found", "video_id", id)
			admResponse.RespondWithError(w, http.StatusNotFound, "Video not found")
			return
		}
		logger.Error("failed to update video", sl.Err(err), "video_id", id)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to update video")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Video updated successfully",
	})
}

// deleteVideo помечает видео как удаленное
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
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
		return
	}

	ctx := r.Context()
	err = h.storage.DeleteVideo(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("video not found", "video_id", id)
			admResponse.RespondWithError(w, http.StatusNotFound, "Video not found")
			return
		}
		logger.Error("failed to delete video", sl.Err(err), "video_id", id)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to delete video")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Video deleted successfully",
	})
}
