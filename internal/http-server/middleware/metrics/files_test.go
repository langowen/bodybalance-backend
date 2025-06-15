package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mw "github.com/langowen/bodybalance-backend/internal/http-server/middleware/metrics"
	projectMetrics "github.com/langowen/bodybalance-backend/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestTempResponseRecorder(t *testing.T) {
	// Тест для TempResponseRecorder
	t.Run("TempResponseRecorder basic functionality", func(t *testing.T) {
		rec := mw.NewTempResponseRecorder()

		// Проверяем начальные значения
		assert.Equal(t, http.StatusOK, rec.Status)
		assert.NotNil(t, rec.Headers)
		assert.NotNil(t, rec.Body)

		// Записываем данные
		rec.WriteHeader(http.StatusCreated)
		content := []byte("test content")
		n, err := rec.Write(content)

		assert.NoError(t, err)
		assert.Equal(t, len(content), n)
		assert.Equal(t, http.StatusCreated, rec.Status)
		assert.Equal(t, "test content", rec.Body.String())

		// Проверяем установку размера файла
		rec.SetFileSize(1024)

		// После установки размера файла, запись не должна изменять Body
		newContent := []byte("new content")
		n, err = rec.Write(newContent)
		assert.NoError(t, err)
		assert.Equal(t, len(newContent), n)
		assert.Equal(t, "test content", rec.Body.String(), "Body should not change after SetFileSize")
	})
}

func TestFileResponseWriter(t *testing.T) {
	// Сбрасываем метрики перед тестом
	projectMetrics.StaticFileRequests.Reset()
	projectMetrics.StaticFileSize.Reset()
	projectMetrics.StaticFileTransferRate.Reset()
	projectMetrics.StaticFileLoadTime.Reset()
	projectMetrics.StaticFileCacheHit.Reset()

	t.Run("FileResponseWriter normal response", func(t *testing.T) {
		// Создаем базовый ResponseWriter
		rr := httptest.NewRecorder()

		// Создаем запрос
		req := httptest.NewRequest("GET", "/img/test.jpg", nil)

		// Создаем FileResponseWriter
		fileType := "image"
		filename := "/img/test.jpg"
		fileSize := int64(1024) // 1KB
		frw := mw.NewFileResponseWriter(rr, req, fileType, filename, fileSize)

		// Записываем заголовок и содержимое
		frw.WriteHeader(http.StatusOK)

		// Создаем тестовые данные (1KB)
		testData := make([]byte, fileSize)
		for i := range testData {
			testData[i] = byte('A' + (i % 26))
		}

		bytesWritten, err := frw.Write(testData)
		assert.NoError(t, err)
		assert.Equal(t, int(fileSize), bytesWritten)

		// Проверяем метрики запросов к файлам
		requestCount := testutil.ToFloat64(projectMetrics.StaticFileRequests.WithLabelValues(fileType, filename, "OK"))
		assert.Equal(t, float64(1), requestCount, "Static file request metric should be 1")

		// Проверяем, что метрики размера файла и скорости передачи были зарегистрированы
		fileSizeMetricCount := testutil.CollectAndCount(projectMetrics.StaticFileSize)
		assert.Greater(t, fileSizeMetricCount, 0, "File size metric should be recorded")

		transferRateMetricCount := testutil.CollectAndCount(projectMetrics.StaticFileTransferRate)
		assert.Greater(t, transferRateMetricCount, 0, "Transfer rate metric should be recorded")

		loadTimeMetricCount := testutil.CollectAndCount(projectMetrics.StaticFileLoadTime)
		assert.Greater(t, loadTimeMetricCount, 0, "Load time metric should be recorded")
	})

	t.Run("FileResponseWriter with 304 cache hit", func(t *testing.T) {
		// Сбрасываем метрику кэш-хитов
		projectMetrics.StaticFileCacheHit.Reset()

		// Создаем базовый ResponseWriter
		rr := httptest.NewRecorder()

		// Создаем запрос с заголовком If-Modified-Since
		req := httptest.NewRequest("GET", "/img/test.jpg", nil)
		req.Header.Set("If-Modified-Since", time.Now().Format(http.TimeFormat))

		// Создаем FileResponseWriter
		fileType := "image"
		filename := "/img/test.jpg"
		fileSize := int64(1024)
		frw := mw.NewFileResponseWriter(rr, req, fileType, filename, fileSize)

		// Возвращаем 304 Not Modified
		frw.WriteHeader(http.StatusNotModified)

		// Проверяем метрику запросов к файлам
		requestCount := testutil.ToFloat64(projectMetrics.StaticFileRequests.WithLabelValues(fileType, filename, "Not Modified"))
		assert.Equal(t, float64(1), requestCount, "Static file 304 request metric should be 1")

		// Проверяем метрику кэш-хитов
		cacheHitCount := testutil.ToFloat64(projectMetrics.StaticFileCacheHit.WithLabelValues(fileType, filename))
		assert.Equal(t, float64(1), cacheHitCount, "Static file cache hit metric should be 1")
	})

	t.Run("FileResponseWriter with error status", func(t *testing.T) {
		// Сбрасываем метрики перед тестом
		projectMetrics.StaticFileRequests.Reset()

		// Создаем базовый ResponseWriter
		rr := httptest.NewRecorder()

		// Создаем запрос
		req := httptest.NewRequest("GET", "/img/nonexistent.jpg", nil)

		// Создаем FileResponseWriter
		fileType := "image"
		filename := "/img/nonexistent.jpg"
		fileSize := int64(0) // Нет файла
		frw := mw.NewFileResponseWriter(rr, req, fileType, filename, fileSize)

		// Возвращаем 404 Not Found
		frw.WriteHeader(http.StatusNotFound)

		// Проверяем метрику запросов к файлам
		requestCount := testutil.ToFloat64(projectMetrics.StaticFileRequests.WithLabelValues(fileType, filename, "Not Found"))
		assert.Equal(t, float64(1), requestCount, "Static file 404 request metric should be 1")
	})
}

// Тест для дополнительного метода из Files middleware
func TestFilesMetricsCalculation(t *testing.T) {
	t.Run("Files metrics calculation", func(t *testing.T) {
		// Сбрасываем метрики перед тестом
		projectMetrics.StaticFileRequests.Reset()
		projectMetrics.StaticFileSize.Reset()
		projectMetrics.StaticFileTransferRate.Reset()
		projectMetrics.StaticFileLoadTime.Reset()

		// Создаем моки
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/img/test.jpg", nil)

		// Создаем FileResponseWriter с известными параметрами
		fileType := "image"
		filename := "/img/test.jpg"
		fileSize := int64(10240) // 10KB

		// Создаем новый FileResponseWriter
		frw := mw.NewFileResponseWriter(rr, req, fileType, filename, fileSize)

		// Записываем заголовок
		frw.WriteHeader(http.StatusOK)

		// Создаем и записываем тестовые данные
		testData := make([]byte, fileSize)
		for i := range testData {
			testData[i] = byte('A' + (i % 26))
		}

		// Записываем данные - это автоматически обновит метрики, если запись
		// достигнет размера файла (согласно реализации в files.go)
		bytesWritten, err := frw.Write(testData)
		assert.NoError(t, err)
		assert.Equal(t, int(fileSize), bytesWritten)

		// Проверяем, что метрики были зарегистрированы
		// Используем CollectAndCount вместо прямой проверки значений
		fileSizeMetricCount := testutil.CollectAndCount(projectMetrics.StaticFileSize)
		assert.Greater(t, fileSizeMetricCount, 0, "File size metric should be recorded")

		transferRateMetricCount := testutil.CollectAndCount(projectMetrics.StaticFileTransferRate)
		assert.Greater(t, transferRateMetricCount, 0, "Transfer rate metric should be recorded")

		loadTimeMetricCount := testutil.CollectAndCount(projectMetrics.StaticFileLoadTime)
		assert.Greater(t, loadTimeMetricCount, 0, "Load time metric should be recorded")
	})
}
