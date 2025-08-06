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

// @Summary Создать нового пользователя
// @Description Добавляет нового пользователя в систему
// @Tags Admin Users
// @Accept json
// @Produce json
// @Param input body dto.UserRequest true "Данные пользователя"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse "Пользователь уже существует"
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/users [post]
func (h *Handler) addUser(w http.ResponseWriter, r *http.Request) {
	const op = "admin.addUser"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	var req dto.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("failed to decode request body", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	user := &admin.Users{
		Username: req.Username,
		ContentType: admin.ContentType{
			ID: req.ContentTypeID,
		},
		IsAdmin:  req.Admin,
		Password: req.Password,
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	result, err := h.service.AddUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrEmptyUsername):
			dto.RespondWithError(w, http.StatusBadRequest, "Введите имя пользователя")
			return
		case errors.Is(err, admin.ErrTypeInvalid):
			dto.RespondWithError(w, http.StatusBadRequest, "Выберите тип контента")
			return
		case errors.Is(err, admin.ErrNotFoundPassword):
			dto.RespondWithError(w, http.StatusBadRequest, "Пароль обязателен для администратора")
			return
		case errors.Is(err, admin.ErrUserAlreadyExists):
			dto.RespondWithError(w, http.StatusConflict, "Пользователь с таким именем уже существует")
			return
		case errors.Is(err, admin.ErrFailedSaveUser):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to add user")
			return
		}
	}

	response := dto.UserResponse{
		ID:            result.ID,
		Username:      result.Username,
		ContentTypeID: result.ContentType.ID,
		ContentType:   result.ContentType.Name,
		Admin:         result.IsAdmin,
		DateCreated:   result.DateCreated,
	}

	dto.RespondWithJSON(w, http.StatusCreated, response)
}

// @Summary Получить пользователя по ID
// @Description Возвращает информацию о пользователе по его идентификатору
// @Tags Admin Users
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/users/{id} [get]
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getUser"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid user ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	user, err := h.service.GetUser(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrUserInvalidID):
			dto.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
			return
		case errors.Is(err, admin.ErrUserNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		case errors.Is(err, admin.ErrFailedGetUser):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
			return
		}
	}

	res := dto.UserResponse{
		ID:            user.ID,
		Username:      user.Username,
		ContentTypeID: user.ContentType.ID,
		ContentType:   user.ContentType.Name,
		Admin:         user.IsAdmin,
		DateCreated:   user.DateCreated,
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Получить всех пользователей
// @Description Возвращает список всех пользователей в системе
// @Tags Admin Users
// @Produce json
// @Success 200 {array} dto.UserResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/users [get]
func (h *Handler) getUsers(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getUsers"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	ctx := logging.ContextWithLogger(r.Context(), logger)

	users, err := h.service.GetUsers(ctx)
	if err != nil {
		if errors.Is(err, admin.ErrUserNotFound) {
			dto.RespondWithError(w, http.StatusNotFound, "No users found")
			return
		}
		logger.Error("failed to get users", sl.Err(err))
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	res := make([]dto.UserResponse, len(users))
	for i, user := range users {
		res[i] = dto.UserResponse{
			ID:            user.ID,
			Username:      user.Username,
			ContentTypeID: user.ContentType.ID,
			ContentType:   user.ContentType.Name,
			Admin:         user.IsAdmin,
			DateCreated:   user.DateCreated,
		}
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Обновить пользователя
// @Description Обновляет информацию о существующем пользователе
// @Tags Admin Users
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Param input body dto.UserRequest true "Новые данные пользователя"
// @Success 200 {object} dto.UserResponse "Пользователь успешно обновлен"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse "Пользователь с таким именем уже существует"
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/users/{id} [put]
func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	const op = "admin.updateUser"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid user ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	user := &admin.Users{
		ID:       id,
		Username: req.Username,
		ContentType: admin.ContentType{
			ID: req.ContentTypeID,
		},
		IsAdmin:  req.Admin,
		Password: req.Password,
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	err = h.service.UpdateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrUserInvalidID):
			dto.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
			return
		case errors.Is(err, admin.ErrEmptyUsername):
			dto.RespondWithError(w, http.StatusBadRequest, "Введите имя пользователя")
			return
		case errors.Is(err, admin.ErrTypeInvalid):
			dto.RespondWithError(w, http.StatusBadRequest, "Выберите тип контента")
			return
		case errors.Is(err, admin.ErrNotFoundPassword):
			dto.RespondWithError(w, http.StatusBadRequest, "Пароль обязателен для администратора")
			return
		case errors.Is(err, admin.ErrUserNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		case errors.Is(err, admin.ErrFailedGetUser):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		case errors.Is(err, admin.ErrUserAlreadyExists):
			dto.RespondWithError(w, http.StatusConflict, "Пользователь с таким именем уже существует")
			return
		}
	}

	res := dto.SuccessResponse{
		ID:      id,
		Message: "User updated successfully",
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Удалить пользователя
// @Description Удаляет пользователя из системы
// @Tags Admin Users
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} dto.SuccessResponse "Пользователь успешно удален"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/users/{id} [delete]
func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	const op = "admin.deleteUser"

	logger := h.logger.With(
		"handler", op,
	)

	ctx := logging.ContextWithLogger(r.Context(), logger)

	userIDParam := chi.URLParam(r, "id")
	if userIDParam == "" {
		logger.Info("no user ID provided")
		dto.RespondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	id, err := strconv.ParseInt(userIDParam, 10, 64)
	if err != nil {
		logger.Error("failed to parse user ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = h.service.DeleteUser(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrUserInvalidID):
			dto.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
			return
		case errors.Is(err, admin.ErrUserNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		case errors.Is(err, admin.ErrFailedGetUser):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
			return
		}
	}

	res := dto.SuccessResponse{
		ID:      id,
		Message: "User deleted successfully",
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}
