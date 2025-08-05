// Package v1 содержит обработчики API версии 1
// @title BodyBalance API
// @version 1.0
// @description API для управления видео-контентом BodyBalance.
// @contact.name Sergei
// @contact.email info@7375.org
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host body.7375.org
// @BasePath /v1
// @schemes https
package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/internal/adapter/storage"
	"github.com/langowen/bodybalance-backend/internal/app"
	"github.com/langowen/bodybalance-backend/internal/entities/api"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/api/v1/dto"
	mwMetrics "github.com/langowen/bodybalance-backend/internal/port/http-server/middleware/metrics"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
)

type Handler struct {
	logger  *logging.Logger
	cfg     *config.Config
	service Service
}

func New(app *app.App) *Handler {
	return &Handler{
		logger:  app.Logger,
		cfg:     app.Cfg,
		service: app.ServiceApi,
	}
}

func (h *Handler) Router(r ...chi.Router) chi.Router {
	var router chi.Router
	if len(r) > 0 {
		router = r[0]
	} else {
		router = chi.NewRouter()
	}

	router.Get("/video_categories", h.getVideosByCategoryAndType)
	router.Get("/video", h.getVideo)
	router.Get("/category", h.getCategoriesByType)
	router.Get("/login", h.checkAccount)
	router.Post("/feedback", h.feedback)

	return router
}

// @Summary Check account existence
// @Description Checks if account with specified username exists and returns type info
// @Tags API v1
// @Produce json
// @Param username query string true "Username to check"
// @Success 200 {object} dto.AccountResponse
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /login [get]
func (h *Handler) checkAccount(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.checkAccount"

	username := r.URL.Query().Get("username")

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
		"username", username,
	)

	ctx := logging.ContextWithLogger(r.Context(), logger)

	account, err := h.service.GetTypeByAccount(ctx, username)
	if err != nil {
		switch {
		case errors.Is(err, api.ErrEmptyUsername):
			dto.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Username is empty")
			return
		case errors.Is(err, api.ErrStorageServerError):
			dto.RespondWithError(w, http.StatusInternalServerError, "Server Error", "Failed to check account")
			return
		case errors.Is(err, storage.ErrAccountNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Account with username %s not found", username))
			return
		}
	}

	mwMetrics.RecordDataSource(r, account.DataSource)

	res := dto.AccountResponse{
		TypeID:   account.ContentType.ID,
		TypeName: account.ContentType.Name,
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Get categories by type
// @Description Returns all categories for specified type, ordered by name
// @Tags API v1
// @Produce json
// @Param type query int true "Type ID"
// @Success 200 {array} dto.CategoryResponse
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /category [get]
func (h *Handler) getCategoriesByType(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.getCategoriesByType"

	contentType := r.URL.Query().Get("type")

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
		"type", contentType,
	)

	ctx := logging.ContextWithLogger(r.Context(), logger)

	categories, err := h.service.GetCategoriesByType(ctx, contentType)
	if err != nil {
		switch {
		case errors.Is(err, api.ErrEmptyTypeID):
			dto.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Content type is empty")
			return
		case errors.Is(err, api.ErrTypeInvalid):
			dto.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Invalid type ID")
			return
		case errors.Is(err, storage.ErrContentTypeNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Content type %s not found", contentType))
			return
		case errors.Is(err, api.ErrStorageServerError):
			dto.RespondWithError(w, http.StatusInternalServerError, "Server Error", "Failed to get categories")
			return
		}
	}

	mwMetrics.RecordDataSource(r, categories[0].DataSource)

	categoriesResponse := make([]dto.CategoryResponse, 0, len(categories))
	for _, category := range categories {
		categoriesResponse = append(categoriesResponse, dto.CategoryResponse{
			ID:     category.ID,
			Name:   category.Name,
			ImgURL: category.ImgURL,
		})
	}

	dto.RespondWithJSON(w, http.StatusOK, categories)
}

// @Summary Get video by ID
// @Description Returns video details by its ID
// @Tags API v1
// @Produce json
// @Param video_id query int true "Video ID"
// @Success 200 {object} dto.VideoResponse
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /video [get]
func (h *Handler) getVideo(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.getVideo"

	videoID := r.URL.Query().Get("video_id")

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
		"video_id", videoID,
	)

	ctx := logging.ContextWithLogger(r.Context(), logger)

	video, err := h.service.GetVideo(ctx, videoID)
	if err != nil {
		switch {
		case errors.Is(err, api.ErrEmptyVideoID):
			dto.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Video id is empty")
			return
		case errors.Is(err, api.ErrInvalidVideoID):
			dto.RespondWithError(w, http.StatusBadRequest, "Bad Request", fmt.Sprintf("Video ID '%s' is not a valid number", videoID))
			return
		case errors.Is(err, storage.ErrVideoNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Video with id %s not found", videoID))
			return
		case errors.Is(err, api.ErrStorageServerError):
			dto.RespondWithError(w, http.StatusInternalServerError, "Server Error", "Failed to get video")
			return
		}
	}

	mwMetrics.RecordDataSource(r, video.DataSource)

	res := dto.VideoResponse{
		ID:          video.ID,
		Name:        video.Name,
		Description: video.Description,
		URL:         video.URL,
		Category:    video.Category.Name,
		ImgURL:      video.ImgURL,
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Get videos by category and type
// @Description Returns videos filtered by type and category, ordered by name
// @Tags API v1
// @Produce json
// @Param type query int true "Type ID"
// @Param category query int true "Category ID"
// @Success 200 {array} dto.VideoResponse
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /video_categories [get]
func (h *Handler) getVideosByCategoryAndType(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.getVideosByCategoryAndType"

	contentType := r.URL.Query().Get("type")
	categoryName := r.URL.Query().Get("category")

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
		"type", contentType,
		"category", categoryName,
	)

	ctx := logging.ContextWithLogger(r.Context(), logger)

	videos, err := h.service.GetVideosByCategoryAndType(ctx, contentType, categoryName)
	if err != nil {
		switch {
		case errors.Is(err, api.ErrEmptyTypeID):
			dto.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Content type is empty")
			return
		case errors.Is(err, api.ErrEmptyCategoryID):
			dto.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Category is empty")
			return
		case errors.Is(err, api.ErrTypeInvalid):
			dto.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Invalid type ID")
			return
		case errors.Is(err, api.ErrCategoryInvalid):
			dto.RespondWithError(w, http.StatusBadRequest, "Bad Request", "Invalid category ID")
			return
		case errors.Is(err, storage.ErrContentTypeNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Content type %s not found", contentType))
			return
		case errors.Is(err, storage.ErrNoCategoriesFound):
			dto.RespondWithError(w, http.StatusNotFound, "Not Found", fmt.Sprintf("Category %s not found", categoryName))
			return
		case errors.Is(err, api.ErrStorageServerError):
			dto.RespondWithError(w, http.StatusInternalServerError, "Server Error", "Failed to get videos")
			return
		case errors.Is(err, storage.ErrVideoNotFound):
			dto.RespondWithError(w, http.StatusNotFound, "Not Found", "No videos found for the specified type and category")
		}
	}

	mwMetrics.RecordDataSource(r, videos[0].DataSource)

	res := make([]dto.VideoResponse, 0, len(videos))
	for _, video := range videos {
		res = append(res, dto.VideoResponse{
			ID:          video.ID,
			URL:         video.URL,
			Name:        video.Name,
			Description: video.Description,
			Category:    video.Category.Name,
			ImgURL:      video.ImgURL,
		})
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Submit feedback
// @Description Saves user feedback to the system. Requires: message, at least one contact method (email or telegram), valid email format if provided, valid telegram handle (@ + 5-32 chars) if provided
// @Tags API v1
// @Accept json
// @Produce json
// @Param feedback body dto.FeedbackResponse true "Feedback data to submit. Required fields: message. At least one of: email (valid format) or telegram (must start with @, 5-32 chars)"
// @Success 200 {string} string "Feedback successfully saved"
// @Failure 400 {object} string "Invalid request: possible reasons - missing required fields, invalid email/telegram format, no contact method provided"
// @Failure 500 {object} string "Server Error"
// @Router /feedback [post]
func (h *Handler) feedback(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.feedback"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	var req dto.FeedbackResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	feedback := api.Feedback{
		Message:  req.Message,
		Email:    req.Email,
		Telegram: req.Telegram,
		Name:     req.Name,
	}

	err := h.service.Feedback(ctx, &feedback)
	if err != nil {
		switch {
		case errors.Is(err, api.ErrEmptyMessage):
			dto.RespondWithError(w, http.StatusBadRequest, "Message is required")
			return
		case errors.Is(err, api.ErrInvalidEmail):
			dto.RespondWithError(w, http.StatusBadRequest, "Invalid email format")
			return
		case errors.Is(err, api.ErrInvalidTelegram):
			dto.RespondWithError(w, http.StatusBadRequest, "Telegram must start with @ and contain 5-32 characters")
			return
		case errors.Is(err, api.ErrEmptyTelegramOrEmail):
			dto.RespondWithError(w, http.StatusBadRequest, "Either email or telegram must be provided")
			return
		case errors.Is(err, api.ErrStorageServerError):
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to save feedback")
			return
		}
	}

	dto.RespondWithJSON(w, http.StatusOK, "Feedback successfully saved")
}
