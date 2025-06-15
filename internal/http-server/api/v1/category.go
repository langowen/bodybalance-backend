package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	mwMetrics "github.com/langowen/bodybalance-backend/internal/http-server/middleware/metrics"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/langowen/bodybalance-backend/internal/storage"
	"github.com/theartofdevel/logging"
	"net/http"
	"strconv"
)

// @Summary Get categories by type
// @Description Returns all categories for specified type, ordered by name
// @Tags API v1
// @Produce json
// @Param type query int true "Type ID"
// @Success 200 {array} response.CategoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /category [get]
func (h *Handler) getCategoriesByType(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.getCategoriesByType"

	contentType := r.URL.Query().Get("type")

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
		"type", contentType,
	)

	if contentType == "" {
		response.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Content type is empty")
		return
	}

	typeID, err := strconv.ParseInt(contentType, 10, 64)
	if err != nil {
		logger.Error("invalid type ID", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid type ID")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	var categories []response.CategoryResponse

	if h.cfg.Redis.Enable {

		categories, err = h.redis.GetCategories(ctx, typeID)
		if err != nil {

			logger.Warn("failed to get categories from redis", sl.Err(err))
		} else if categories != nil {

			mwMetrics.RecordDataSource(r, mwMetrics.SourceRedis)

			logger.Debug("categories fetched from redis cache")

			response.RespondWithJSON(w, http.StatusOK, categories)

			return
		}
	}

	categories, err = h.storage.GetCategories(ctx, typeID)
	if err != nil {
		if errors.Is(err, storage.ErrContentTypeNotFound) {
			response.RespondWithError(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Content type %d not found", typeID))
			return
		}

		logger.Error("failed to get categories from DB", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Server Error", "Failed to get categories")
		return
	}

	mwMetrics.RecordDataSource(r, mwMetrics.SourceSQL)

	if h.cfg.Redis.Enable && categories != nil {
		go func(ctx context.Context, typeID int64, categories []response.CategoryResponse) {

			if err := h.redis.SetCategories(ctx, typeID, categories, h.cfg.Redis.CacheTTL); err != nil {
				logger.Warn("failed to cache categories in redis", sl.Err(err))
			}
		}(ctx, typeID, categories)
	}

	response.RespondWithJSON(w, http.StatusOK, categories)
}
