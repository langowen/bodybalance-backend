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

// GET /v1/video_categories?type={type}&category={category}
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
		response.RespondWithError(w, http.StatusBadRequest, "Category is empty")
	}

	if contentType == "" {
		logger.Error("Content type is empty")
		response.RespondWithError(w, http.StatusBadRequest, "Content type is empty")
	}

	// Создаем новый контекст с логгером
	ctx := logging.ContextWithLogger(r.Context(), logger)

	videos, err := h.storage.GetVideosByCategoryAndType(ctx, contentType, categoryName)
	switch {
	case errors.Is(err, storage.ErrContentTypeNotFound):
		logger.Warn("content type not found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Content type not found",
			fmt.Sprintf("Content type '%s' does not exist", contentType))
		return

	case errors.Is(err, storage.ErrNoCategoriesFound):
		logger.Warn("no categories found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Category not found",
			fmt.Sprintf("Category '%s' does not exist", categoryName))
		return

	case errors.Is(err, storage.ErrVideoNotFound):
		logger.Warn("video not found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Video not found",
			fmt.Sprintf("no videos found for content type '%s' and category '%s'",
				contentType, categoryName))
		return

	case err != nil:
		logger.Error("Failed to get videos", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Failed to get videos")
		return
	}

	response.RespondWithJSON(w, http.StatusOK, videos)
}
