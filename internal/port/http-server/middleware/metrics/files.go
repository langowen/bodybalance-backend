package metrics

import (
	"bytes"
	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/internal/port/metrics"
	"net/http"
	"time"
)

// TempResponseRecorder временный рекордер ответа для предварительной проверки файла
type TempResponseRecorder struct {
	Headers   http.Header
	Body      *bytes.Buffer
	Status    int
	fileSize  int64
	fileFound bool
}

func NewTempResponseRecorder() *TempResponseRecorder {
	return &TempResponseRecorder{
		Headers: make(http.Header),
		Body:    new(bytes.Buffer),
		Status:  http.StatusOK,
	}
}

func (trr *TempResponseRecorder) Header() http.Header {
	return trr.Headers
}

func (trr *TempResponseRecorder) Write(b []byte) (int, error) {
	if trr.fileFound {
		// Если файл уже найден, нам не нужно писать его содержимое
		return len(b), nil
	}
	return trr.Body.Write(b)
}

func (trr *TempResponseRecorder) WriteHeader(status int) {
	trr.Status = status
}

// Метод для установки размера файла
func (trr *TempResponseRecorder) SetFileSize(size int64) {
	trr.fileSize = size
	trr.fileFound = true
}

// FileResponseWriter отслеживает количество отправленных байтов
type FileResponseWriter struct {
	http.ResponseWriter
	bytesWritten  int64
	status        int
	wroteHeader   bool
	startTime     time.Time
	fileSize      int64
	fileType      string
	filename      string
	originalReq   *http.Request
	loadCompleted bool
}

func NewFileResponseWriter(w http.ResponseWriter, r *http.Request, fileType, filename string, fileSize int64) *FileResponseWriter {
	return &FileResponseWriter{
		ResponseWriter: w,
		startTime:      time.Now(),
		fileType:       fileType,
		filename:       filename,
		originalReq:    r,
		fileSize:       fileSize,
	}
}

func (frw *FileResponseWriter) WriteHeader(code int) {
	if !frw.wroteHeader {
		frw.status = code
		frw.wroteHeader = true
		frw.ResponseWriter.WriteHeader(code)

		// Записываем метрику запроса к файлу
		metrics.StaticFileRequests.WithLabelValues(
			frw.fileType,
			frw.filename,
			http.StatusText(code),
		).Inc()

		if code == http.StatusNotModified {
			// Если файл не изменился (304), значит это кэш-хит
			metrics.StaticFileCacheHit.WithLabelValues(frw.fileType, frw.filename).Inc()
		}
	}
}

func (frw *FileResponseWriter) Write(b []byte) (int, error) {
	if !frw.wroteHeader {
		frw.WriteHeader(http.StatusOK)
	}

	n, err := frw.ResponseWriter.Write(b)
	frw.bytesWritten += int64(n)

	// Если это последняя запись и передача завершена
	if err == nil && frw.bytesWritten == frw.fileSize && !frw.loadCompleted {
		frw.loadCompleted = true
		duration := time.Since(frw.startTime).Seconds()

		// Записываем метрики окончания загрузки
		if duration > 0 && frw.bytesWritten > 0 {
			// Время загрузки
			metrics.StaticFileLoadTime.WithLabelValues(frw.fileType, frw.filename).Observe(duration)

			// Размер файла
			metrics.StaticFileSize.WithLabelValues(frw.fileType, frw.filename).Observe(float64(frw.bytesWritten))

			// Скорость передачи в байтах/секунду
			transferRate := float64(frw.bytesWritten) / duration
			metrics.StaticFileTransferRate.WithLabelValues(frw.fileType, frw.filename).Observe(transferRate)
		}
	}

	return n, err
}

// WrapVideoHandler оборачивает обработчик видео для сбора метрик
func WrapVideoHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := chi.URLParam(r, "filename")

		// Для запросов с If-None-Match или If-Modified-Since
		// передаем управление напрямую основному обработчику, но регистрируем метрики
		if r.Header.Get("If-None-Match") != "" || r.Header.Get("If-Modified-Since") != "" {
			fileResponseWriter := NewFileResponseWriter(w, r, "video", filename, 0)
			next.ServeHTTP(fileResponseWriter, r)

			// Если статус был 304 Not Modified, увеличиваем счетчик кэш-хитов
			if fileResponseWriter.status == http.StatusNotModified {
				metrics.StaticFileCacheHit.WithLabelValues("video", filename).Inc()
			}
			return
		}

		// Для обычных запросов выполняем предварительную проверку
		tempResponseRecorder := NewTempResponseRecorder()
		next.ServeHTTP(tempResponseRecorder, r)

		// Если файл не найден или другая ошибка, просто перенаправляем ответ
		if tempResponseRecorder.Status != http.StatusOK {
			for k, v := range tempResponseRecorder.Headers {
				w.Header()[k] = v
			}
			w.WriteHeader(tempResponseRecorder.Status)
			w.Write(tempResponseRecorder.Body.Bytes())

			// Записываем метрику запроса с ошибкой
			metrics.StaticFileRequests.WithLabelValues(
				"video",
				filename,
				http.StatusText(tempResponseRecorder.Status),
			).Inc()
			return
		}

		// Файл найден, отправляем с метриками
		fileSize := tempResponseRecorder.fileSize
		fileResponseWriter := NewFileResponseWriter(w, r, "video", filename, fileSize)

		// Копируем заголовки из предварительной проверки
		for k, v := range tempResponseRecorder.Headers {
			fileResponseWriter.Header()[k] = v
		}

		// Передаем управление исходному обработчику с оберткой для метрик
		next.ServeHTTP(fileResponseWriter, r)
	}
}

// WrapImgHandler оборачивает обработчик изображений для сбора метрик
func WrapImgHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := chi.URLParam(r, "filename")

		// Для запросов с If-None-Match или If-Modified-Since
		// передаем управление напрямую основному обработчику, но регистрируем метрики
		if r.Header.Get("If-None-Match") != "" || r.Header.Get("If-Modified-Since") != "" {
			fileResponseWriter := NewFileResponseWriter(w, r, "image", filename, 0)
			next.ServeHTTP(fileResponseWriter, r)

			// Если статус был 304 Not Modified, увеличиваем счетчик кэш-хитов
			if fileResponseWriter.status == http.StatusNotModified {
				metrics.StaticFileCacheHit.WithLabelValues("image", filename).Inc()
			}
			return
		}

		// Для обычных запросов выполняем предварительную проверку
		tempResponseRecorder := NewTempResponseRecorder()
		next.ServeHTTP(tempResponseRecorder, r)

		// Если файл не найден или другая ошибка, просто перенаправляем ответ
		if tempResponseRecorder.Status != http.StatusOK {
			for k, v := range tempResponseRecorder.Headers {
				w.Header()[k] = v
			}
			w.WriteHeader(tempResponseRecorder.Status)
			w.Write(tempResponseRecorder.Body.Bytes())

			// Записываем метрику запроса с ошибкой
			metrics.StaticFileRequests.WithLabelValues(
				"image",
				filename,
				http.StatusText(tempResponseRecorder.Status),
			).Inc()
			return
		}

		// Файл найден, отправляем с метриками
		fileSize := tempResponseRecorder.fileSize
		fileResponseWriter := NewFileResponseWriter(w, r, "image", filename, fileSize)

		// Копируем заголовки из предварительной проверки
		for k, v := range tempResponseRecorder.Headers {
			fileResponseWriter.Header()[k] = v
		}

		// Передаем управление исходному обработчику с оберткой для метрик
		next.ServeHTTP(fileResponseWriter, r)
	}
}
