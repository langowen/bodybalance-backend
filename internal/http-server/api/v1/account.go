package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/langowen/bodybalance-backend/internal/storage"
	"github.com/theartofdevel/logging"
	"net/http"
	"time"
)

// @Summary Check account existence
// @Description Checks if account with specified username exists and returns type info
// @Tags API v1
// @Produce json
// @Param username query string true "Username to check"
// @Success 200 {object} response.AccountResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /login [get]
func (h *Handler) checkAccount(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.checkAccountType"
	const cacheTTL = time.Hour * 24

	username := r.URL.Query().Get("username")

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

	ctx := logging.ContextWithLogger(r.Context(), logger)

	// Пытаемся получить данные из кэша
	cachedAccount, err := h.redis.GetAccount(ctx, username)
	if err != nil {
		logger.Warn("redis get error", sl.Err(err))
	}

	if cachedAccount != nil {
		logger.Debug("serving from cache", "account_type", cachedAccount.TypeName)

		// Устанавливаем токен аутентификации в cookie
		if err := h.SetAuthCookie(w, username, cachedAccount.TypeID, cachedAccount.TypeName); err != nil {
			logger.Error("failed to set auth cookie", sl.Err(err))
			// Даже если не смогли установить cookie, все равно возвращаем данные
		}

		response.RespondWithJSON(w, http.StatusOK, cachedAccount)
		return
	}

	// Данных нет в кэше - запрашиваем из основного хранилища
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

	// Сохраняем данные в кэш
	go func() {
		ctx := context.Background()

		if err := h.redis.SetAccount(ctx, username, &account, cacheTTL); err != nil {
			logger.Warn("failed to set account cache", sl.Err(err))
		}
	}()

	// Устанавливаем токен аутентификации в cookie
	if err := h.SetAuthCookie(w, username, account.TypeID, account.TypeName); err != nil {
		logger.Error("failed to set auth cookie", sl.Err(err))
		// Даже если не смогли установить cookie, все равно возвращаем данные
	}

	response.RespondWithJSON(w, http.StatusOK, account)
}
