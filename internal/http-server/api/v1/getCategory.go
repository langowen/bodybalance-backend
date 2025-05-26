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

// @Summary Get categories by  type
// @Description Returns all categories for specified type, order by name
// @Tags Categories
// @Accept  json
// @Produce  json
// @Param type query int true "Type id (e.g. '1', '2')"
// @Success 200 {array} response.CategoryResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v1/category [get]
// GET /v1/category?type={id}
func (h *Handler) getCategoriesByType(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.getCategoriesByType"

	contentType := r.URL.Query().Get("type")

	// Создаем логгер с дополнительными полями
	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
		"type", contentType,
	)

	if contentType == "" {
		response.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Content type is empty")
		return
	}

	// Создаем новый контекст с логгером
	ctx := logging.ContextWithLogger(r.Context(), logger)

	categories, err := h.storage.GetCategories(ctx, contentType)
	switch {
	case errors.Is(err, storage.ErrContentTypeNotFound):
		logger.Warn("content type not found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Not Found", "Content type not found",
			fmt.Sprintf("Content type '%s' does not exist", contentType))
		return

	case errors.Is(err, storage.ErrNoCategoriesFound):
		logger.Warn("no categories found", sl.Err(err))
		response.RespondWithError(w, http.StatusNotFound, "Not Found", "Category not found",
			fmt.Sprintf("Category '%s' does not exist", contentType))
		return

	case err != nil:
		logger.Error("Failed to get categories", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	response.RespondWithJSON(w, http.StatusOK, categories)
}
