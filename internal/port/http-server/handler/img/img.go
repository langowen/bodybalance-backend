package img

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/api/v1/dto"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/middleware/metrics"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
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
			dto.RespondWithError(w, http.StatusForbidden, "Forbidden")
			return
		}

		file, err := os.Open(cleanPath)
		if err != nil {
			logger.Error("File not found", "filename", filename, sl.Err(err))
			dto.RespondWithError(w, http.StatusNotFound, "Not Found", err.Error())
			return
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			logger.Error("Failed to get file info", "filename", filename, sl.Err(err))
			dto.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
			return
		}

		if tempRecorder, ok := w.(*metrics.TempResponseRecorder); ok {
			tempRecorder.SetFileSize(fileInfo.Size())
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("ETag", fmt.Sprintf(`"%x-%x"`, fileInfo.Size(), fileInfo.ModTime().UnixNano()))

		http.ServeContent(w, r, filename, fileInfo.ModTime(), file)
	}
}
