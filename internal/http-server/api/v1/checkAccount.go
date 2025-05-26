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

// @Summary Check account existence
// @Description Checks if account with specified username exists and return type id, type name
// @Tags Auth
// @Accept  json
// @Produce  json
// @Produce  text/plain
// @Param username query string true "Username to check (e.g. 'base')"
// @Success 200 {object} response.AccountResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/login [get]
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
		response.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Username is empty")
		return
	}

	// Создаем новый контекст с логгером
	ctx := logging.ContextWithLogger(r.Context(), logger)

	account, err := h.storage.CheckAccount(ctx, username)
	switch {
	case errors.Is(err, storage.ErrAccountNotFound):
		logger.Warn("Username not found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Not Found", "Username not found",
			fmt.Sprintf("Username '%s' does not exist", username))
		return

	case err != nil:
		logger.Error("Failed to check account", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	response.RespondWithJSON(w, http.StatusOK, account)
}
