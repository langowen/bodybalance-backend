package admin

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin/dto"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
	"strconv"
)

// @Summary Создать новую категорию
// @Description Добавляет новую категорию в систему
// @Tags Admin Categories
// @Accept json
// @Produce json
// @Param input body dto.CategoryRequest true "Данные категории"
// @Success 201 {object} dto.CategoryResponse "Категория успешно создана"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/category [post]
func (h *Handler) addCategory(w http.ResponseWriter, r *http.Request) {
	const op = "admin.addCategory"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	ctx := logging.ContextWithLogger(r.Context(), logger)

	var req dto.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	catReq := admin.Category{
		Name:        req.Name,
		ImgURL:      req.ImgURL,
		ContentType: make([]admin.ContentType, len(req.TypeIDs)),
	}

	for i, typeID := range req.TypeIDs {
		catReq.ContentType[i].ID = typeID
	}

	category, err := h.service.AddCategory(ctx, &catReq)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrEmptyName):
			dto.RespondWithError(w, http.StatusBadRequest, "Введите название категории")
			return
		case errors.Is(err, admin.ErrEmptyImgURL):
			dto.RespondWithError(w, http.StatusBadRequest, "Выберите превью для категории")
			return
		case errors.Is(err, admin.ErrEmptyTypeIDs):
			dto.RespondWithError(w, http.StatusBadRequest, "Выберите хотя бы один тип контента")
			return
		case errors.Is(err, admin.ErrInvalidImgFormat):
			dto.RespondWithError(w, http.StatusBadRequest, "Недопустимый формат имени файла превью")
			return
		case errors.Is(err, admin.ErrSuspiciousContent):
			dto.RespondWithError(w, http.StatusBadRequest, "Подозрительный контент в URL превью")
			return
		default:
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to add category")
			return
		}
	}

	res := dto.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		ImgURL:      category.ImgURL,
		Types:       make([]dto.TypeResponse, len(category.ContentType)),
		DateCreated: category.CreatedAt,
	}

	for i, contentType := range category.ContentType {
		res.Types[i] = dto.TypeResponse{
			ID:   contentType.ID,
			Name: contentType.Name,
		}
	}

	dto.RespondWithJSON(w, http.StatusCreated, res)
}

// @Summary Получить категорию по ID
// @Description Возвращает информацию о категории по её идентификатору
// @Tags Admin Categories
// @Produce json
// @Param id path int true "ID категории"
// @Success 200 {object} dto.CategoryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/category/{id} [get]
func (h *Handler) getCategory(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getCategory"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid category ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	category, err := h.service.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, admin.ErrCategoryNotFound) {
			dto.RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		}
		logger.Error("failed to get category", sl.Err(err), "category_id", id)
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get category")
		return
	}

	res := dto.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		ImgURL:      category.ImgURL,
		Types:       make([]dto.TypeResponse, len(category.ContentType)),
		DateCreated: category.CreatedAt,
	}

	for i, contentType := range category.ContentType {
		res.Types[i] = dto.TypeResponse{
			ID:   contentType.ID,
			Name: contentType.Name,
		}
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Получить все категории
// @Description Возвращает список всех категорий в системе
// @Tags Admin Categories
// @Produce json
// @Success 200 {array} dto.CategoryResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/category [get]
func (h *Handler) getCategories(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getCategories"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	ctx := logging.ContextWithLogger(r.Context(), logger)
	categories, err := h.service.GetCategories(ctx)
	if err != nil {
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get categories")
		return
	}

	res := make([]dto.CategoryResponse, len(categories))
	for i, category := range categories {
		res[i] = dto.CategoryResponse{
			ID:          category.ID,
			Name:        category.Name,
			ImgURL:      category.ImgURL,
			Types:       make([]dto.TypeResponse, len(category.ContentType)),
			DateCreated: category.CreatedAt,
		}
		for j, contentType := range category.ContentType {
			res[i].Types[j] = dto.TypeResponse{
				ID:   contentType.ID,
				Name: contentType.Name,
			}
		}
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Обновить категорию
// @Description Обновляет информацию о существующей категории
// @Tags Admin Categories
// @Accept json
// @Produce json
// @Param id path int true "ID категории"
// @Param input body dto.CategoryRequest true "Новые данные категории"
// @Success 200 {object} dto.CategoryResponse "Категория успешно обновлена"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/category/{id} [put]
func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	const op = "admin.updateCategory"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid category ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid category ID", "category_id", idStr)
		return
	}

	var req dto.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	catReq := admin.Category{
		ID:          id,
		Name:        req.Name,
		ImgURL:      req.ImgURL,
		ContentType: make([]admin.ContentType, len(req.TypeIDs)),
	}
	for i, typeID := range req.TypeIDs {
		catReq.ContentType[i].ID = typeID
	}

	err = h.service.UpdateCategory(ctx, id, &catReq)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrEmptyName):
			dto.RespondWithError(w, http.StatusBadRequest, "Введите название категории")
			return
		case errors.Is(err, admin.ErrEmptyImgURL):
			dto.RespondWithError(w, http.StatusBadRequest, "Выберите превью для категории")
			return
		case errors.Is(err, admin.ErrEmptyTypeIDs):
			dto.RespondWithError(w, http.StatusBadRequest, "Выберите хотя бы один тип контента")
			return
		case errors.Is(err, admin.ErrInvalidImgFormat):
			dto.RespondWithError(w, http.StatusBadRequest, "Недопустимый формат имени файла превью")
			return
		case errors.Is(err, admin.ErrSuspiciousContent):
			dto.RespondWithError(w, http.StatusBadRequest, "Подозрительный контент в URL превью")
			return
		case errors.Is(err, admin.ErrCategoryNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		default:
			logger.Error("failed to update category", sl.Err(err), "category_id", id)
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to update category")
			return
		}
	}

	res := dto.SuccessResponse{
		ID:      id,
		Message: "Category updated successfully",
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Удалить категорию
// @Description Удаляет категорию из системы
// @Tags Admin Categories
// @Produce json
// @Param id path int true "ID категории"
// @Success 200 {object} dto.SuccessResponse "Категория успешно удалена"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/category/{id} [delete]
func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	const op = "admin.deleteCategory"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid category ID", sl.Err(err), "category_id", idStr)
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	err = h.service.DeleteCategory(ctx, id)
	if err != nil {
		if errors.Is(err, admin.ErrCategoryNotFound) {
			dto.RespondWithError(w, http.StatusNotFound, "Category not found")
			return
		}
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to delete category")
		return
	}

	res := dto.SuccessResponse{
		ID:      id,
		Message: "Category deleted successfully",
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}
