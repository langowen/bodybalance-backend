package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
	"strconv"
	"strings"
)

// @Summary Создать новую категорию
// @Description Добавляет новую категорию в систему
// @Tags Admin Categories
// @Accept json
// @Produce json
// @Param input body admResponse.CategoryRequest true "Данные категории"
// @Success 201 {object} admResponse.CategoryResponse "Категория успешно создана"
// @Failure 400 {object} admResponse.ErrorResponse
// @Failure 500 {object} admResponse.ErrorResponse
// @security AdminAuth
// @Router /admin/category [post]
func (h *Handler) addCategory(w http.ResponseWriter, r *http.Request) {
	const op = "admin.addCategory"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	var req admResponse.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if !h.validCat(&req, w, logger) {
		return
	}

	ctx := r.Context()
	category, err := h.storage.AddCategory(ctx, &req)
	if err != nil {
		logger.Error("failed to add category", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to add category")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusCreated, category)
}

// @Summary Получить категорию по ID
// @Description Возвращает информацию о категории по её идентификатору
// @Tags Admin Categories
// @Produce json
// @Param id path int true "ID категории"
// @Success 200 {object} admResponse.CategoryResponse
// @Failure 400 {object} admResponse.ErrorResponse
// @Failure 404 {object} admResponse.ErrorResponse
// @Failure 500 {object} admResponse.ErrorResponse
// @security AdminAuth
// @Router /admin/category/{id} [get]
func (h *Handler) getCategory(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getCategory"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid category ID", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	ctx := r.Context()
	category, err := h.storage.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("category not found", "category_id", id)
			admResponse.RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		}
		logger.Error("failed to get category", sl.Err(err), "category_id", id)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to get category")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, category)
}

// @Summary Получить все категории
// @Description Возвращает список всех категорий в системе
// @Tags Admin Categories
// @Produce json
// @Success 200 {array} admResponse.CategoryResponse
// @Failure 500 {object} admResponse.ErrorResponse
// @security AdminAuth
// @Router /admin/category [get]
func (h *Handler) getCategories(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getCategories"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	ctx := r.Context()
	categories, err := h.storage.GetCategories(ctx)
	if err != nil {
		logger.Error("failed to get categories", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to get categories")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, categories)
}

// @Summary Обновить категорию
// @Description Обновляет информацию о существующей категории
// @Tags Admin Categories
// @Accept json
// @Produce json
// @Param id path int true "ID категории"
// @Param input body admResponse.CategoryRequest true "Новые данные категории"
// @Success 200 {object} object{id=int64,message=string} "Категория успешно обновлена"
// @Failure 400 {object} admResponse.ErrorResponse
// @Failure 404 {object} admResponse.ErrorResponse
// @Failure 500 {object} admResponse.ErrorResponse
// @security AdminAuth
// @Router /admin/category/{id} [put]
func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	const op = "admin.updateCategory"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid category ID", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	var req admResponse.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if !h.validCat(&req, w, logger) {
		return
	}

	ctx := r.Context()

	err = h.storage.UpdateCategory(ctx, id, &req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("category not found", "category_id", id)
			admResponse.RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		}
		logger.Error("failed to update category", sl.Err(err), "category_id", id)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to update category")
		return
	}

	if h.cfg.Redis.Enable == true {
		go h.removeCategoryCache(id)
	}

	admResponse.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Category updated successfully",
	})
}

// @Summary Удалить категорию
// @Description Удаляет категорию из системы
// @Tags Admin Categories
// @Produce json
// @Param id path int true "ID категории"
// @Success 200 {object} object{id=int64,message=string} "Категория успешно удалена"
// @Failure 400 {object} admResponse.ErrorResponse
// @Failure 404 {object} admResponse.ErrorResponse
// @Failure 500 {object} admResponse.ErrorResponse
// @security AdminAuth
// @Router /admin/category/{id} [delete]
func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	const op = "admin.deleteCategory"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid category ID", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	ctx := r.Context()

	err = h.storage.DeleteCategory(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("category not found", "category_id", id)
			admResponse.RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		}
		logger.Error("failed to delete category", sl.Err(err), "category_id", id)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to delete category")
		return
	}

	if h.cfg.Redis.Enable == true {
		go h.removeCategoryCache(id)
	}

	admResponse.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Category deleted successfully",
	})
}

// removeCategoryCache удаляет записи из кэша при обновлении категорий
func (h *Handler) removeCategoryCache(id int64) {
	const op = "admin.removeCategoryCache"

	logger := h.logger.With(
		"op", op,
		"category_id", id)

	ctx := context.Background()

	category, err := h.storage.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug("category not found", "category_id", id)
		}
		logger.Debug("failed to get category", sl.Err(err), "category_id", id)
		return
	}

	if category == nil {
		logger.Debug("category is nil", "category_id", id)
		return
	}

	for _, contentType := range category.Types {
		err := h.redis.DeleteCategories(ctx, contentType.ID)
		if err != nil {
			logger.Warn("failed to invalidate cache for type", sl.Err(err), "type_id", contentType.ID)
		}

		err = h.redis.DeleteVideosByCategoryAndType(
			ctx, contentType.ID, category.ID)
		if err != nil {
			h.logger.Warn("failed to invalidate videos cache", sl.Err(err))
		}
	}
}

// validCat проверят данные входящего запроса на их валидность
func (h *Handler) validCat(req *admResponse.CategoryRequest, w http.ResponseWriter, logger *logging.Logger) bool {
	switch {
	case req.Name == "":
		logger.Warn("empty required Name")
		admResponse.RespondWithError(w, http.StatusBadRequest, "Введите название категории")
		return false
	case req.ImgURL == "":
		logger.Warn("empty required ImgURL")
		admResponse.RespondWithError(w, http.StatusBadRequest, "Выберите превью для категории")
		return false
	case len(req.TypeIDs) == 0:
		logger.Warn("empty required TypeIDs")
		admResponse.RespondWithError(w, http.StatusBadRequest, "Выберите хотя бы один тип контента")
		return false
	}

	// Проверка ImgURL
	if !validFilePattern.MatchString(req.ImgURL) {
		logger.Warn("invalid file format in ImgURL", "imgurl", req.ImgURL)
		admResponse.RespondWithError(w, http.StatusBadRequest, "Недопустимый формат имени файла превью")
		return false
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(req.ImgURL, pattern) {
			logger.Warn("suspicious pattern in ImgURL", "imgurl", req.ImgURL, "pattern", pattern)
			admResponse.RespondWithError(w, http.StatusBadRequest, "Недопустимые символы в имени файла превью")
			return false
		}
	}

	return true
}
