package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
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

	cachedCategories, err := h.redis.GetCategories(ctx, typeID)
	if err != nil {
		logger.Warn("redis get error", sl.Err(err))
	}

	if cachedCategories != nil {
		logger.Debug("serving from cache")
		response.RespondWithJSON(w, http.StatusOK, cachedCategories)
		return
	}

	categories, err := h.storage.GetCategories(ctx, typeID)
	switch {
	case errors.Is(err, storage.ErrContentTypeNotFound):
		logger.Warn("content type not found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Not Found", "Content type not found",
			fmt.Sprintf("Content type '%d' does not exist", typeID))
		return

	case errors.Is(err, storage.ErrNoCategoriesFound):
		logger.Warn("no categories found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Not Found", "Category not found",
			fmt.Sprintf("Category '%d' does not exist", typeID))
		return

	case err != nil:
		logger.Error("Failed to get categories", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	go func() {
		ctx := context.Background()
		if err := h.redis.SetCategories(ctx, typeID, categories, h.cfg.Redis.CacheTTL); err != nil {
			logger.Warn("failed to set cache", sl.Err(err))
		}
	}()

	response.RespondWithJSON(w, http.StatusOK, categories)
}
