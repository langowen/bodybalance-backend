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
	"strings"
)

func (h *Handler) addUser(w http.ResponseWriter, r *http.Request) {
	const op = "admin.addUser"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	var req admResponse.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Username == "" || req.Password == "" {
		logger.Error("empty required fields")
		admResponse.RespondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	ctx := r.Context()
	user, err := h.storage.AddUser(ctx, req)
	if err != nil {
		logger.Error("failed to add user", sl.Err(err))

		// Проверяем, является ли ошибка ошибкой дубликата
		if strings.Contains(err.Error(), "user already exists") {
			admResponse.RespondWithError(w, http.StatusConflict, "duplicate key")
			return
		}

		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to add user")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"id":              user.ID,
		"username":        user.Username,
		"content_type_id": user.ContentTypeID,
		"content_type":    user.ContentType,
		"admin":           user.Admin,
		"date_created":    user.DateCreated,
		"message":         "User added successfully",
	})
}

// getUser возвращает пользователя по ID
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getUser"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid user ID", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	ctx := r.Context()
	user, err := h.storage.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("user not found", "user_id", id)
			admResponse.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		logger.Error("failed to get user", sl.Err(err), "user_id", id)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, user)
}

// getUsers возвращает всех пользователей
func (h *Handler) getUsers(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getUsers"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	ctx := r.Context()
	users, err := h.storage.GetUsers(ctx)
	if err != nil {
		logger.Error("failed to get users", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, users)
}

// updateUser обновляет данные пользователя
func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	const op = "admin.updateUser"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid user ID", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req admResponse.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	ctx := r.Context()
	err = h.storage.UpdateUser(ctx, id, req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("user not found", "user_id", id)
			admResponse.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		logger.Error("failed to update user", sl.Err(err), "user_id", id)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "User updated successfully",
	})
}

// deleteUser помечает пользователя как удаленного
func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	const op = "admin.deleteUser"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid user ID", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	ctx := r.Context()
	err = h.storage.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("user not found", "user_id", id)
			admResponse.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		logger.Error("failed to delete user", sl.Err(err), "user_id", id)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "User deleted successfully",
	})
}
