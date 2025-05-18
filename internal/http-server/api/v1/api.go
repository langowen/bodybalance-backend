package api

import (
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/storage"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/langowen/bodybalance-backend/internal/storage/postgres"
	"github.com/theartofdevel/logging"
)

type Handler struct {
	logger  *logging.Logger
	storage *postgres.Storage
	cfg     *config.Config
}

type CategoryResponse struct {
	ID        int             `json:"id"`         // ID из БД
	Category  string          `json:"category"`   // Название категории
	VideoList []storage.Video `json:"videoItems"` // Список видео
}

func New(router *chi.Mux, logger *logging.Logger, storage *postgres.Storage, cfg *config.Config) {
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

	h.logger = h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	contentType := r.URL.Query().Get("type")
	categoryName := r.URL.Query().Get("category")

	videos, err := h.storage.GetVideosByCategoryAndType(r.Context(), contentType, categoryName)
	if err != nil {
		h.logger.Error("Failed to get videos",
			"type", contentType,
			"category", categoryName,
			sl.Err(err))

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(videos); err != nil {
		h.logger.Error("Failed to encode response", sl.Err(err))
	}
}

// GET /v1/category?type={type}
func (h *Handler) getCategoriesByType(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.getCategoriesByType"

	h.logger = h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	contentType := r.URL.Query().Get("type")

	categories, err := h.storage.GetCategoriesWithVideos(r.Context(), contentType)
	if err != nil {
		h.logger.Error("Failed to get categories with videos",
			"type", contentType,
			sl.Err(err))
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

	h.logger = h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	username := r.URL.Query().Get("username")
	contentType := r.URL.Query().Get("type")

	isValid, err := h.storage.CheckAccountType(r.Context(), username, contentType)
	if err != nil {
		h.logger.Error("Failed to check account", sl.Err(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(isValid); err != nil {
		h.logger.Error("Failed to encode response", sl.Err(err))
	}
}
