package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
	"os"
	"path/filepath"
)

func ServeVideoFile(cfg *config.Config, logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.video.get.download"

		logger = logger.With(
			"op", op,
			"request_id", middleware.GetReqID(r.Context()),
		)

		filename := chi.URLParam(r, "filename")
		if filename == "" {
			http.Error(w, "Filename not specified", http.StatusBadRequest)
			return
		}

		// Безопасное формирование пути к файлу
		filePath := filepath.Join(cfg.Media.VideoPath, filename)

		// Проверяем, что файл существует
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			logger.Error("File not found", "filename", filename, sl.Err(err))
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Устанавливаем правильный Content-Type
		w.Header().Set("Content-Type", "video/mp4")

		// Отдаем файл
		http.ServeFile(w, r, filePath)
	}
}
