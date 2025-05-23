package video

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func ServeVideoFile(cfg *config.Config, logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.video.ServeVideoFile"

		logger = logger.With(
			"op", op,
			"request_id", middleware.GetReqID(r.Context()),
		)

		filename := chi.URLParam(r, "filename")
		filePath := filepath.Join(cfg.Media.VideoPath, filename)

		if !strings.HasPrefix(filepath.Clean(filePath), cfg.Media.VideoPath) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			logger.Error("File not found", "filename", filename, sl.Err(err))
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			logger.Error("File stat error", "filename", filename, sl.Err(err))
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, filename, stat.ModTime(), file)
	}
}
