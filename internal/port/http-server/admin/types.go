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

// @Summary Создать новый тип
// @Description Добавляет новый тип контента в систему
// @Tags Admin Types
// @Accept json
// @Produce json
// @Param input body dto.TypeRequest true "Данные типа"
// @Success 201 {object} dto.TypeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/type [post]
func (h *Handler) addType(w http.ResponseWriter, r *http.Request) {
	const op = "admin.addType"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	var req dto.TypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	contentType := &admin.ContentType{
		Name: req.Name,
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	result, err := h.service.AddType(ctx, contentType)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrTypeNameEmpty):
			dto.RespondWithError(w, http.StatusBadRequest, "Name is required")
			return
		case errors.Is(err, admin.ErrFailedSaveType):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to add type")
			return
		}
	}

	response := dto.TypeResponse{
		ID:          result.ID,
		Name:        result.Name,
		DateCreated: result.DateCreated,
	}

	dto.RespondWithJSON(w, http.StatusCreated, response)
}

// @Summary Получить тип по ID
// @Description Возвращает информацию о типе контента по его идентификатору
// @Tags Admin Types
// @Produce json
// @Param id path int true "ID типа"
// @Success 200 {object} dto.TypeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/type/{id} [get]
func (h *Handler) getType(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getType"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid type ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid type ID")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	contentType, err := h.service.GetType(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrTypeInvalid):
			dto.RespondWithError(w, http.StatusBadRequest, "Invalid type ID")
			return
		case errors.Is(err, admin.ErrTypeNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "Type not found")
			return
		case errors.Is(err, admin.ErrFailedGetType):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get type")
			return
		}
	}

	res := dto.TypeResponse{
		ID:          contentType.ID,
		Name:        contentType.Name,
		DateCreated: contentType.DateCreated,
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Получить все типы
// @Description Возвращает список всех типов контента в системе
// @Tags Admin Types
// @Produce json
// @Success 200 {array} dto.TypeResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/type [get]
func (h *Handler) getTypes(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getTypes"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	ctx := logging.ContextWithLogger(r.Context(), logger)

	types, err := h.service.GetTypes(ctx)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrTypeNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "No types found")
			return
		case errors.Is(err, admin.ErrFailedGetType):
			logger.Error("failed to get types", sl.Err(err))
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get types")
			return
		}
	}

	res := make([]dto.TypeResponse, len(types))
	for i, contentType := range types {
		res[i] = dto.TypeResponse{
			ID:          contentType.ID,
			Name:        contentType.Name,
			DateCreated: contentType.DateCreated,
		}
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Обновить тип
// @Description Обновляет информацию о существующем типе контента
// @Tags Admin Types
// @Accept json
// @Produce json
// @Param id path int true "ID типа"
// @Param input body dto.TypeRequest true "Новые данные типа"
// @Success 200 {object} dto.SuccessResponse "Тип успешно обновлен"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/type/{id} [put]
func (h *Handler) updateType(w http.ResponseWriter, r *http.Request) {
	const op = "admin.updateType"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid type ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid type ID")
		return
	}

	var req dto.TypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	contentType := &admin.ContentType{
		ID:   id,
		Name: req.Name,
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	err = h.service.UpdateType(ctx, contentType)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrTypeInvalid):
			dto.RespondWithError(w, http.StatusBadRequest, "Invalid type ID")
			return
		case errors.Is(err, admin.ErrTypeNameEmpty):
			dto.RespondWithError(w, http.StatusBadRequest, "Name is required")
			return
		case errors.Is(err, admin.ErrTypeNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "Type not found")
			return
		case errors.Is(err, admin.ErrFailedSaveType):
			logger.Error("failed to update type", sl.Err(err))
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to update type")
			return
		}
	}

	res := dto.SuccessResponse{
		ID:      id,
		Message: "Type updated successfully",
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Удалить тип
// @Description Удаляет тип контента из системы
// @Tags Admin Types
// @Produce json
// @Param id path int true "ID типа"
// @Success 200 {object} dto.SuccessResponse "Тип успешно удален"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/type/{id} [delete]
func (h *Handler) deleteType(w http.ResponseWriter, r *http.Request) {
	const op = "admin.deleteType"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid type ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid type ID")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	err = h.service.DeleteType(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrTypeInvalid):
			dto.RespondWithError(w, http.StatusBadRequest, "Invalid type ID")
			return
		case errors.Is(err, admin.ErrTypeNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "Type not found")
			return
		case errors.Is(err, admin.ErrFailedSaveType):
			logger.Error("failed to delete type", sl.Err(err))
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to delete type")
			return
		}
	}

	res := dto.SuccessResponse{
		ID:      id,
		Message: "Type deleted successfully",
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}
