package video

import (
	"fmt"
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
	"time"
)

// ServeVideoFile
// @Summary Serve video file
// @Description Streams video file by filename
// @Tags Files
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

		start := time.Now()

		logger = logger.With(
			"op", op,
			"request_id", middleware.GetReqID(r.Context()),
		)

		filename := chi.URLParam(r, "filename")
		filePath := filepath.Join(cfg.Media.VideoPatch, filename)

		cleanPath := filepath.Clean(filePath)
		if !strings.HasPrefix(cleanPath, cfg.Media.VideoPatch) {
			response.RespondWithError(w, http.StatusForbidden, "Forbidden")
			return
		}

		file, err := os.Open(cleanPath)
		if err != nil {
			logger.Error("File not found", "filename", filename, sl.Err(err))
			response.RespondWithError(w, http.StatusNotFound, "Not Found", err.Error())
			return
		}
		defer file.Close()

		// Логируем начало обработки
		logger.Debug("start serving video",
			"filename", filename,
			"method", r.Method,
			"range_header", r.Header.Get("Range"),
			"eTag", r.Header.Get("ETag"))

		// Замер времени чтения файла
		readStart := time.Now()
		stat, err := file.Stat()
		if err != nil {
			logger.Error("File stat error", "filename", filename, sl.Err(err))
			response.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
			return
		}
		readDuration := time.Since(readStart)

		etag := fmt.Sprintf("\"%x-%x\"", stat.Size(), stat.ModTime().Unix())
		w.Header().Set("ETag", etag)

		if match := r.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		serveStart := time.Now()
		http.ServeContent(w, r, filename, stat.ModTime(), file)
		serveDuration := time.Since(serveStart)

		// Логируем результаты
		logger.Debug("video served",
			"filename", filename,
			"file_size", stat.Size(),
			"read_duration", readDuration.String(),
			"server_duration", serveDuration.String(),
			"total_duration", time.Since(start).String(),
			"range_header", r.Header.Get("Range"),
			"eTag", r.Header.Get("ETag"),
			"status", getStatusFromResponse(w))
	}
}

func getStatusFromResponse(w http.ResponseWriter) int {
	if rw, ok := w.(interface{ Status() int }); ok {
		return rw.Status()
	}
	return http.StatusOK
}
