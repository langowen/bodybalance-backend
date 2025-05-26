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

// @Summary Get videos by category and type
// @Description Returns videos filtered by type and category, order by name
// @Tags Videos
// @Accept  json
// @Produce  json
// @Param type query int true "Type id (e.g. '1')"
// @Param category query int true "Category id(e.g. '1')"
// @Success 200 {array} response.VideoResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/video_categories [get]
// GET /v1/video_categories?type={id}&category={id}
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
