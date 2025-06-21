package img

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/http-server/middleware/metrics"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ServeImgFile
// @Summary Serve img file
// @Description Send img file by filename
// @Tags Files
// @Accept  json
// @Produce  json
// @Param filename path string true "Video filename (e.g. 'shee_video.jpg')"
// @Success 200 {file} file
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /img/{filename} [get]
// GET /img/{filename}
func ServeImgFile(cfg *config.Config, logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.images.ServeImgFile"

		logger = logger.With(
			"op", op,
			"request_id", middleware.GetReqID(r.Context()),
		)

		filename := chi.URLParam(r, "filename")

		filePath := filepath.Join(cfg.Media.ImagesPatch, filename)

		cleanPath := filepath.Clean(filePath)
		if !strings.HasPrefix(cleanPath, cfg.Media.ImagesPatch) {
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

		fileInfo, err := file.Stat()
		if err != nil {
			logger.Error("Failed to get file info", "filename", filename, sl.Err(err))
			response.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
			return
		}

		if tempRecorder, ok := w.(*metrics.TempResponseRecorder); ok {
			tempRecorder.SetFileSize(fileInfo.Size())
			return
		}

		// Генерируем ETag на основе имени файла и времени последнего изменения
		etag := fmt.Sprintf(`"%s-%d"`, filename, fileInfo.ModTime().UnixNano())
		w.Header().Set("ETag", etag)

		// Проверяем If-None-Match заголовок
		if match := r.Header.Get("If-None-Match"); match != "" && match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		http.ServeContent(w, r, filename, fileInfo.ModTime(), file)
	}
}
