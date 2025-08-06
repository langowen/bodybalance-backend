package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin/dto"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
)

// @Summary Добавить новое видео
// @Description Создает новую запись видео в системе
// @Tags Admin Videos
// @Accept json
// @Produce json
// @Param input body dto.VideoRequest true "Данные видео"
// @Success 201 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/video [post]
func (h *Handler) addVideo(w http.ResponseWriter, r *http.Request) {
	const op = "admin.addVideo"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	var req dto.VideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	video := &admin.Video{
		Name:        req.Name,
		URL:         req.URL,
		ImgURL:      req.ImgURL,
		Description: req.Description,
		Categories:  make([]admin.Category, len(req.CategoryIDs)),
	}
	for i, catID := range req.CategoryIDs {
		video.Categories[i] = admin.Category{
			ID: catID,
		}
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	videoID, err := h.service.AddVideo(ctx, video)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrVideoInvalidName):
			dto.RespondWithError(w, http.StatusBadRequest, "Введите название видео")
			return
		case errors.Is(err, admin.ErrVideoInvalidURL):
			dto.RespondWithError(w, http.StatusBadRequest, "Введите URL видео")
			return
		case errors.Is(err, admin.ErrVideoInvalidImgURL):
			dto.RespondWithError(w, http.StatusBadRequest, "Введите URL изображения видео")
			return
		case errors.Is(err, admin.ErrVideoInvalidCategory):
			dto.RespondWithError(w, http.StatusBadRequest, "Выберите категории для видео")
			return
		case errors.Is(err, admin.ErrVideoURLPattern):
			dto.RespondWithError(w, http.StatusBadRequest, "Неверный формат URL видео")
			return
		case errors.Is(err, admin.ErrVideoSuspiciousPattern):
			dto.RespondWithError(w, http.StatusBadRequest, "Подозрительный URL видео")
			return
		case errors.Is(err, admin.ErrVideoImgPattern):
			dto.RespondWithError(w, http.StatusBadRequest, "Неверный формат URL изображения видео")
			return
		case errors.Is(err, admin.ErrVideoImgSuspiciousPattern):
			dto.RespondWithError(w, http.StatusBadRequest, "Подозрительный URL изображения видео")
			return
		case errors.Is(err, admin.ErrVideoSaveFailed):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to save video")
			return
		case errors.Is(err, admin.ErrCategoryNotFound):
			dto.RespondWithError(w, http.StatusBadRequest, "One or more categories not found")
			return
		}
	}

	res := dto.SuccessResponse{
		ID:      videoID,
		Message: "Video added successfully",
	}

	dto.RespondWithJSON(w, http.StatusCreated, res)
}

// @Summary Получить видео по ID
// @Description Возвращает информацию о конкретном видео
// @Tags Admin Videos
// @Produce json
// @Param id path int true "ID видео"
// @Success 200 {object} dto.VideoResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/video/{id} [get]
func (h *Handler) getVideo(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getVideo"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid video ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	video, err := h.service.GetVideo(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrVideoInvalidID):
			dto.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
			return
		case errors.Is(err, admin.ErrVideoNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "Video not found")
			return
		case errors.Is(err, admin.ErrFailedGetVideo):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get video")
			return
		}
	}

	res := dto.VideoResponse{
		ID:          video.ID,
		Name:        video.Name,
		URL:         video.URL,
		ImgURL:      video.ImgURL,
		Description: video.Description,
		Categories:  make([]dto.CategoryResponse, len(video.Categories)),
	}
	for i, cat := range video.Categories {
		res.Categories[i] = dto.CategoryResponse{
			ID:          cat.ID,
			Name:        cat.Name,
			ImgURL:      cat.ImgURL,
			DateCreated: cat.CreatedAt,
		}
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Получить список всех видео
// @Description Возвращает список всех видео в системе
// @Tags Admin Videos
// @Produce json
// @Success 200 {array} dto.VideoResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/video [get]
func (h *Handler) getVideos(w http.ResponseWriter, r *http.Request) {
	const op = "admin.getVideos"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	ctx := logging.ContextWithLogger(r.Context(), logger)

	videos, err := h.service.GetVideos(ctx)
	if err != nil {
		if errors.Is(err, admin.ErrVideoNotFound) {
			dto.RespondWithError(w, http.StatusNotFound, "No videos found")
			return
		}
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to get videos")
		return
	}

	res := make([]dto.VideoResponse, len(videos))
	for i, video := range videos {
		res[i] = dto.VideoResponse{
			ID:          video.ID,
			Name:        video.Name,
			URL:         video.URL,
			ImgURL:      video.ImgURL,
			Description: video.Description,
			Categories:  make([]dto.CategoryResponse, len(video.Categories)),
			DateCreated: video.DateCreated,
		}
		for j, cat := range video.Categories {
			res[i].Categories[j] = dto.CategoryResponse{
				ID:          cat.ID,
				Name:        cat.Name,
				ImgURL:      cat.ImgURL,
				DateCreated: cat.CreatedAt,
			}
		}
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Обновить данные видео
// @Description Обновляет информацию о существующем видео
// @Tags Admin Videos
// @Accept json
// @Produce json
// @Param id path int true "ID видео"
// @Param input body dto.VideoRequest true "Новые данные видео"
// @Success 200 {object} dto.SuccessResponse "Видео успешно обновлено"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/video/{id} [put]
func (h *Handler) updateVideo(w http.ResponseWriter, r *http.Request) {
	const op = "admin.updateVideo"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid video ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	var req dto.VideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	video := &admin.Video{
		ID:          id,
		Name:        req.Name,
		URL:         req.URL,
		ImgURL:      req.ImgURL,
		Description: req.Description,
		Categories:  make([]admin.Category, len(req.CategoryIDs)),
	}
	for i, catID := range req.CategoryIDs {
		video.Categories[i] = admin.Category{
			ID: catID,
		}
	}

	err = h.service.UpdateVideo(ctx, video)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrVideoInvalidID):
			dto.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
			return
		case errors.Is(err, admin.ErrVideoInvalidName):
			dto.RespondWithError(w, http.StatusBadRequest, "Введите название видео")
			return
		case errors.Is(err, admin.ErrVideoInvalidURL):
			dto.RespondWithError(w, http.StatusBadRequest, "Введите URL видео")
			return
		case errors.Is(err, admin.ErrVideoInvalidImgURL):
			dto.RespondWithError(w, http.StatusBadRequest, "Введите URL изображения видео")
			return
		case errors.Is(err, admin.ErrVideoInvalidCategory):
			dto.RespondWithError(w, http.StatusBadRequest, "Выберите категории для видео")
			return
		case errors.Is(err, admin.ErrVideoURLPattern):
			dto.RespondWithError(w, http.StatusBadRequest, "Неверный формат URL видео")
			return
		case errors.Is(err, admin.ErrVideoSuspiciousPattern):
			dto.RespondWithError(w, http.StatusBadRequest, "Подозрительный URL видео")
			return
		case errors.Is(err, admin.ErrVideoImgPattern):
			dto.RespondWithError(w, http.StatusBadRequest, "Неверный формат URL изображения видео")
			return
		case errors.Is(err, admin.ErrVideoImgSuspiciousPattern):
			dto.RespondWithError(w, http.StatusBadRequest, "Подозрительный URL изображения видео")
			return
		case errors.Is(err, admin.ErrVideoUpdateFailed):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to update video")
			return
		case errors.Is(err, admin.ErrVideoNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "Video not found")
			return
		}
	}

	res := dto.SuccessResponse{
		ID:      id,
		Message: "Video updated successfully",
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Удалить видео
// @Description Помечает видео как удаленное в системе
// @Tags Admin Videos
// @Produce json
// @Param id path int true "ID видео"
// @Success 200 {object} object{id=int64,message=string} "Видео успешно удалено"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @security AdminAuth
// @Router /admin/video/{id} [delete]
func (h *Handler) deleteVideo(w http.ResponseWriter, r *http.Request) {
	const op = "admin.deleteVideo"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error("invalid video ID", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	err = h.service.DeleteVideo(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrVideoInvalidID):
			dto.RespondWithError(w, http.StatusBadRequest, "Invalid video ID")
			return
		case errors.Is(err, admin.ErrVideoNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "Video not found")
			return
		case errors.Is(err, admin.ErrVideoDeleteFailed):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to delete video")
			return
		}
	}

	res := dto.SuccessResponse{
		ID:      id,
		Message: "Video deleted successfully",
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}
