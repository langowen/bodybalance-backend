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
	"net/http"
	"strconv"
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

	if req.Name == "" || len(req.TypeIDs) == 0 {
		logger.Error("empty required fields")
		admResponse.RespondWithError(w, http.StatusBadRequest, "Name and at least one type_id are required")
		return
	}

	ctx := r.Context()
	category, err := h.storage.AddCategory(ctx, req)
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

	if req.Name == "" || len(req.TypeIDs) == 0 {
		logger.Error("empty required fields")
		admResponse.RespondWithError(w, http.StatusBadRequest, "Name and at least one type_id are required")
		return
	}

	ctx := r.Context()

	// Получаем текущие типы категории перед обновлением (для инвалидации кэша)
	category, err := h.storage.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug("category not found", "category_id", id)
		}
		logger.Debug("failed to get category", sl.Err(err), "category_id", id)
	}

	err = h.storage.UpdateCategory(ctx, id, req)
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

	// Инвалидация кэша для всех затронутых типов контента
	go func() {
		ctx := context.Background() // используем новый контекст для фоновой операции

		// Инвалидируем кэш для старых типов категории
		for _, catType := range category.Types {
			if err := h.redis.DeleteCategories(ctx, strconv.FormatFloat(catType.ID, 'f', 0, 64)); err != nil {
				logger.Warn("failed to invalidate cache for old type",
					sl.Err(err), "type_id", catType.ID)
			}
		}

		// Инвалидируем кэш для новых типов категории
		for _, typeID := range req.TypeIDs {
			if err := h.redis.DeleteCategories(ctx, strconv.FormatInt(typeID, 10)); err != nil {
				logger.Warn("failed to invalidate cache for new type",
					sl.Err(err), "type_id", typeID)
			}
		}
	}()

	go func() {
		// Инвалидируем кэш для всех связанных тип+категория комбинаций
		for _, contentType := range category.Types {
			if err := h.redis.DeleteVideosByCategoryAndType(
				context.Background(),
				strconv.FormatFloat(contentType.ID, 'f', 0, 64),
				strconv.FormatFloat(category.ID, 'f', 0, 64),
			); err != nil {
				h.logger.Warn("failed to invalidate videos cache", sl.Err(err))
			}
		}
	}()

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

	// Получаем текущие типы категории перед обновлением (для инвалидации кэша)
	category, err := h.storage.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug("category not found", "category_id", id)
		}
		logger.Debug("failed to get category", sl.Err(err), "category_id", id)
	}

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

	// Инвалидация кэша для всех связанных типов контента
	go func() {
		ctx := context.Background() // используем новый контекст для фоновой операции
		for _, catType := range category.Types {
			if err := h.redis.DeleteCategories(ctx, strconv.FormatFloat(catType.ID, 'f', 0, 64)); err != nil {
				logger.Warn("failed to invalidate cache for type",
					sl.Err(err), "type_id", catType.ID)
			}
		}
	}()

	go func() {
		// Инвалидируем кэш для всех связанных тип+категория комбинаций
		for _, contentType := range category.Types {
			if err := h.redis.DeleteVideosByCategoryAndType(
				context.Background(),
				strconv.FormatFloat(contentType.ID, 'f', 0, 64),
				strconv.FormatFloat(category.ID, 'f', 0, 64),
			); err != nil {
				h.logger.Warn("failed to invalidate videos cache", sl.Err(err))
			}
		}
	}()

	admResponse.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Category deleted successfully",
	})
}
