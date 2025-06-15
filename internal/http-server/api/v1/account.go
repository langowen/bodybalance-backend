package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	mwMetrics "github.com/langowen/bodybalance-backend/internal/http-server/middleware/metrics"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/langowen/bodybalance-backend/internal/storage"
	"github.com/theartofdevel/logging"
	"net/http"
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

	if h.cfg.Redis.Enable == true {
		cachedAccount, err := h.redis.GetAccount(ctx, username)
		if err != nil {
			logger.Warn("redis get error", sl.Err(err))
		}

		if cachedAccount != nil {
			logger.Debug("serving from cache", "account_type", cachedAccount.TypeName)

			mwMetrics.RecordDataSource(r, mwMetrics.SourceRedis)

			response.RespondWithJSON(w, http.StatusOK, cachedAccount)
			return
		}
	}

	account, err := h.storage.CheckAccount(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrAccountNotFound) {
			response.RespondWithError(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Account with username %s not found", username))
			return
		}

		logger.Error("storage get error", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Server Error", "Failed to check account")
		return
	}

	mwMetrics.RecordDataSource(r, mwMetrics.SourceSQL)

	if h.cfg.Redis.Enable {
		go func(ctx context.Context, username string, acc *response.AccountResponse) {
			if err := h.redis.SetAccount(ctx, username, acc, h.cfg.Redis.CacheTTL); err != nil {
				logger.Warn("failed to cache account in redis", sl.Err(err))
			}
		}(ctx, username, account)
	}

	response.RespondWithJSON(w, http.StatusOK, account)
}
