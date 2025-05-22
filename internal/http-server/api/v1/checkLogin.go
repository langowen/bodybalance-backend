package v1

import (
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
)

// GET /v1/login?username={username}
func (h *Handler) checkAccount(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.checkAccountType"

	username := r.URL.Query().Get("username")

	// Создаем логгер с дополнительными полями
	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
		"username", username,
	)

	// Создаем новый контекст с логгером
	ctx := logging.ContextWithLogger(r.Context(), logger)

	isValid, err := h.storage.CheckAccount(ctx, username)
	if err != nil {
		logger.Error("Failed to check account", sl.Err(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(isValid); err != nil {
		logger.Error("Failed to encode response", sl.Err(err))
	}
}
