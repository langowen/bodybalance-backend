package video

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ServeVideoFile
// @Summary Serve video file
// @Description Streams video file by filename
// @Tags Videos
// @Accept  json
// @Produce  json
// @Param filename path string true "Video filename (e.g. 'Sheya_baza.mp4')"
// @Success 200 {file} file
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /video/{filename} [get]
// GET /video/{filename}
func ServeVideoFile(cfg *config.Config, logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.video.ServeVideoFile"

		logger = logger.With(
			"op", op,
			"request_id", middleware.GetReqID(r.Context()),
		)

		filename := chi.URLParam(r, "filename")
		filePath := filepath.Join(cfg.Media.VideoPatch, filename)

		if !strings.HasPrefix(filepath.Clean(filePath), cfg.Media.VideoPatch) {
			response.RespondWithError(w, http.StatusForbidden, "Forbidden")
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			logger.Error("File not found", "filename", filename, sl.Err(err))
			response.RespondWithError(w, http.StatusNotFound, "Not Found", err.Error())
			return
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			logger.Error("File stat error", "filename", filename, sl.Err(err))
			response.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
			return
		}

		http.ServeContent(w, r, filename, stat.ModTime(), file)
	}
}
