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

// GET /v1/video?video_id={id}
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
		response.RespondWithError(w, http.StatusBadRequest, "Video id is empty")
	}

	// Создаем новый контекст с логгером
	ctx := logging.ContextWithLogger(r.Context(), logger)

	videos, err := h.storage.GetVideo(ctx, videoID)
	switch {
	case errors.Is(err, storage.ErrVideoNotFound):
		logger.Warn("video not found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Video not found",
			fmt.Sprintf("Video '%s' does not exist", videoID))
		return

	case err != nil:
		logger.Error("Failed to get video", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Failed to get video")
		return
	}

	response.RespondWithJSON(w, http.StatusOK, videos)
}
