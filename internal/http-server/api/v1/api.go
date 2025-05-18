package api

import (
	"encoding/json"
	"github.com/langowen/bodybalance-backend/internal/config"
	"net/http"
	"os"
	"path/filepath"

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

	router.Get("/date/video/{filename}", h.serveVideoFile)
}

type CategoryResponse struct {
	ContentType string   `json:"content_type"`
	Categories  []string `json:"categories"`
}

func (h *Handler) serveVideoFile(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		http.Error(w, "Filename not specified", http.StatusBadRequest)
		return
	}

	// Безопасное формирование пути к файлу
	filePath := filepath.Join(h.cfg.Media.VideoPath, filename)

	// Проверяем, что файл существует
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.logger.Error("File not found", "filename", filename, sl.Err(err))
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Устанавливаем правильный Content-Type
	w.Header().Set("Content-Type", "video/mp4")

	// Отдаем файл
	http.ServeFile(w, r, filePath)
}

// GET /v1/video?type={type}&category={category}
func (h *Handler) getVideosByCategoryAndType(w http.ResponseWriter, r *http.Request) {
	contentType := r.URL.Query().Get("type")
	category := r.URL.Query().Get("category")

	videos, err := h.storage.GetVideosByCategoryAndType(r.Context(), contentType, category)
	if err != nil {
		h.logger.Error("Failed to get videos", sl.Err(err))
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
	contentType := r.URL.Query().Get("type")

	categories, err := h.storage.GetCategoriesByContentType(r.Context(), contentType)
	if err != nil {
		h.logger.Error("Failed to get categories", sl.Err(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := CategoryResponse{
		ContentType: contentType,
		Categories:  categories,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", sl.Err(err))
	}
}

// GET /v1/login?username={username}&type={type}
func (h *Handler) checkAccountType(w http.ResponseWriter, r *http.Request) {
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
