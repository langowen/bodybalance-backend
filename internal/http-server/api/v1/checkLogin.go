package v1

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
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

	if username == "" {
		logger.Error("Username is empty")
		response.RespondWithError(w, http.StatusBadRequest, "Username is empty")
		return
	}

	// Создаем новый контекст с логгером
	ctx := logging.ContextWithLogger(r.Context(), logger)

	isValid, err := h.storage.CheckAccount(ctx, username)
	if err != nil {
		logger.Error("Failed to check account", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.RespondWithJSON(w, http.StatusOK, isValid)
}
