package api

import (
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/storage"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
)

type Handler struct {
	logger  *logging.Logger
	storage storage.ApiStorage
	cfg     *config.Config
}

type CategoryResponse struct {
	ID        int             `json:"id"`         // ID из БД
	Category  string          `json:"category"`   // Название категории
	VideoList []storage.Video `json:"videoItems"` // Список видео
}

func New(router *chi.Mux, logger *logging.Logger, storage storage.ApiStorage, cfg *config.Config) {
	h := &Handler{
		logger:  logger,
		storage: storage,
		cfg:     cfg,
	}

	router.Route("/v1", func(r chi.Router) {
		r.Get("/video", h.getVideosByCategoryAndType)
		r.Get("/category", h.getCategoriesByType)
		r.Get("/login", h.checkAccountType)
	})
}

// GET /v1/video?type={type}&category={category}
func (h *Handler) getVideosByCategoryAndType(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.getVideosByCategoryAndType"

	contentType := r.URL.Query().Get("type")
	categoryName := r.URL.Query().Get("category")

	// Создаем логгер с дополнительными полями
	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
		"type", contentType,
		"category", categoryName,
	)

	// Создаем новый контекст с логгером
	ctx := logging.ContextWithLogger(r.Context(), logger)

	videos, err := h.storage.GetVideosByCategoryAndType(ctx, contentType, categoryName)
	if err != nil {
		logger.Error("Failed to get videos", sl.Err(err))

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(videos); err != nil {
		logger.Error("Failed to encode response", sl.Err(err))
	}
}

// GET /v1/category?type={type}
func (h *Handler) getCategoriesByType(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.getCategoriesByType"

	contentType := r.URL.Query().Get("type")

	// Создаем логгер с дополнительными полями
	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
		"type", contentType,
	)

	// Создаем новый контекст с логгером
	ctx := logging.ContextWithLogger(r.Context(), logger)

	categories, err := h.storage.GetCategoriesWithVideos(ctx, contentType)
	if err != nil {
		logger.Error("Failed to get categories with videos", sl.Err(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(categories); err != nil {
		h.logger.Error("Failed to encode response", sl.Err(err))
	}
}

// GET /v1/login?username={username}&type={type}
func (h *Handler) checkAccountType(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.checkAccountType"

	username := r.URL.Query().Get("username")
	contentType := r.URL.Query().Get("type")

	// Создаем логгер с дополнительными полями
	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
		"type", contentType,
		"username", username,
	)

	// Создаем новый контекст с логгером
	ctx := logging.ContextWithLogger(r.Context(), logger)

	isValid, err := h.storage.CheckAccountType(ctx, username, contentType)
	if err != nil {
		logger.Error("Failed to check account", sl.Err(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(isValid); err != nil {
		logger.Error("Failed to encode response", sl.Err(err))
	}
}
