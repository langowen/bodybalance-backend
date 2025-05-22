package v1

import (
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
)

// GET /v1/video?type={type}&category={category}
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

	// Создаем новый контекст с логгером
	ctx := logging.ContextWithLogger(r.Context(), logger)

	videos, err := h.storage.GetVideosByCategoryAndType(ctx, contentType, categoryName)
	if err != nil {
		logger.Error("Failed to get videos", sl.Err(err))

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(videos); err != nil {
		logger.Error("Failed to encode response", sl.Err(err))
	}
}
