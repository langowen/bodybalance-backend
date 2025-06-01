package admin

import (
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

// addType добавляет новый тип
func (h *Handler) addType(w http.ResponseWriter, r *http.Request) {
	const op = "admin.addType"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	var req admResponse.TypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Name == "" {
		logger.Error("empty required field")
		admResponse.RespondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	ctx := r.Context()
	typeID, err := h.storage.AddType(ctx, req)
	if err != nil {
		logger.Error("failed to add type", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to add type")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      typeID,
		"message": "Type added successfully",
	})
}

// getType возвращает тип по его ID
func (h *Handler) getType(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getType"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid type ID", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid type ID")
		return
	}

	ctx := r.Context()
	contentType, err := h.storage.GetType(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("type not found", "type_id", id)
			admResponse.RespondWithError(w, http.StatusNotFound, "Type not found")
			return
		}
		logger.Error("failed to get type", sl.Err(err), "type_id", id)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to get type")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, contentType)
}

// getTypes возвращает все типы
func (h *Handler) getTypes(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getTypes"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	ctx := r.Context()
	types, err := h.storage.GetTypes(ctx)
	if err != nil {
		logger.Error("failed to get types", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to get types")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, types)
}

// updateType обновляет данные типа
func (h *Handler) updateType(w http.ResponseWriter, r *http.Request) {
	const op = "admin.updateType"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid type ID", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid type ID")
		return
	}

	var req admResponse.TypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	ctx := r.Context()
	err = h.storage.UpdateType(ctx, id, req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("type not found", "type_id", id)
			admResponse.RespondWithError(w, http.StatusNotFound, "Type not found")
			return
		}
		logger.Error("failed to update type", sl.Err(err), "type_id", id)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to update type")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Type updated successfully",
	})
}

// deleteType помечает тип как удаленный
func (h *Handler) deleteType(w http.ResponseWriter, r *http.Request) {
	const op = "admin.deleteType"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid type ID", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid type ID")
		return
	}

	ctx := r.Context()
	err = h.storage.DeleteType(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("type not found", "type_id", id)
			admResponse.RespondWithError(w, http.StatusNotFound, "Type not found")
			return
		}
		logger.Error("failed to delete type", sl.Err(err), "type_id", id)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to delete type")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Type deleted successfully",
	})
}
