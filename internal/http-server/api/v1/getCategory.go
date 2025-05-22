package v1

import (
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
)

// GET /v1/category?type={type}
func (h *Handler) getCategoriesByType(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.getCategoriesByType"

	contentType := r.URL.Query().Get("type")

	// Создаем логгер с дополнительными полями
	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
		"type", contentType,
	)

	// Создаем новый контекст с логгером
	ctx := logging.ContextWithLogger(r.Context(), logger)

	categories, err := h.storage.GetCategoriesWithVideos(ctx, contentType)
	if err != nil {
		logger.Error("Failed to get categories with videos", sl.Err(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(categories); err != nil {
		h.logger.Error("Failed to encode response", sl.Err(err))
	}
}
