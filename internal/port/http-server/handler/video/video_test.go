package video

import (
	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/logdiscart"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestServeVideoFile_Success(t *testing.T) {
	// Создаем временную директорию для тестовых видео
	tmpDir := t.TempDir()
	testVideoPath := filepath.Join(tmpDir, "test.mp4")

	// Создаем тестовый файл
	testVideo := []byte("fake video content")
	err := os.WriteFile(testVideoPath, testVideo, 0644)
	assert.NoError(t, err)

	// Настраиваем конфигурацию
	cfg := &config.Config{
		Media: config.Media{
			VideoPatch: tmpDir,
		},
	}

	logger := logdiscart.NewDiscardLogger()

	// Создаем роутер с нужным параметром
	r := chi.NewRouter()
	r.Get("/video/{filename}", ServeVideoFile(cfg, logger))

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/video/test.mp4", nil)
	rec := httptest.NewRecorder()

	// Выполняем запрос
	r.ServeHTTP(rec, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, testVideo, rec.Body.Bytes())
}

func TestServeVideoFile_NotFound(t *testing.T) {
	// Создаем временную директорию для тестовых видео
	tmpDir := t.TempDir()

	// Настраиваем конфигурацию
	cfg := &config.Config{
		Media: config.Media{
			VideoPatch: tmpDir,
		},
	}

	logger := logdiscart.NewDiscardLogger()

	// Создаем роутер с нужным параметром
	r := chi.NewRouter()
	r.Get("/video/{filename}", ServeVideoFile(cfg, logger))

	// Создаем запрос к несуществующему файлу
	req := httptest.NewRequest(http.MethodGet, "/video/nonexistent.mp4", nil)
	rec := httptest.NewRecorder()

	// Выполняем запрос
	r.ServeHTTP(rec, req)

	// Проверяем результат
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestServeVideoFile_PathTraversal(t *testing.T) {
	// Создаем временную директорию для тестовых видео
	tmpDir := t.TempDir()

	// Создаем вложенную директорию для тестирования path traversal
	nestedDir := filepath.Join(tmpDir, "nested")
	err := os.MkdirAll(nestedDir, 0755)
	assert.NoError(t, err)

	// Создаем тестовый файл во вложенной директории
	validFile := filepath.Join(nestedDir, "valid.mp4")
	err = os.WriteFile(validFile, []byte("valid video"), 0644)
	assert.NoError(t, err)

	// Создаем секретный файл в корневой директории
	secretFile := filepath.Join(tmpDir, "secret.txt")
	err = os.WriteFile(secretFile, []byte("secret data"), 0644)
	assert.NoError(t, err)

	// Настраиваем конфигурацию так, чтобы разрешено было только обращение к nested директории
	cfg := &config.Config{
		Media: config.Media{
			VideoPatch: nestedDir,
		},
	}

	logger := logdiscart.NewDiscardLogger()

	// Создаем роутер с нужным параметром
	r := chi.NewRouter()
	r.Get("/video/{filename}", ServeVideoFile(cfg, logger))

	// Проверяем корректный доступ к файлу
	req := httptest.NewRequest(http.MethodGet, "/video/valid.mp4", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Используем относительный путь для попытки выхода за пределы директории
	traversalPath := "../secret.txt"

	req = httptest.NewRequest(http.MethodGet, "/video/"+traversalPath, nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Проверяем, что запрос либо отклонен с Forbidden, либо не найден
	// Оба варианта допустимы с точки зрения безопасности
	code := rec.Code
	if code != http.StatusForbidden && code != http.StatusNotFound {
		t.Errorf("Expected status code 403 or 404, but got %d", code)
	}
}

func TestServeVideoFile_EmptyFilename(t *testing.T) {
	// Создаем временную директорию для тестовых видео
	tmpDir := t.TempDir()

	// Настраиваем конфигурацию
	cfg := &config.Config{
		Media: config.Media{
			VideoPatch: tmpDir,
		},
	}

	logger := logdiscart.NewDiscardLogger()

	// Создаем роутер с нужным параметром
	r := chi.NewRouter()
	r.Get("/video/{filename}", ServeVideoFile(cfg, logger))

	// Создаем запрос с пустым именем файла
	req := httptest.NewRequest(http.MethodGet, "/video/", nil)
	rec := httptest.NewRecorder()

	// Выполняем запрос
	r.ServeHTTP(rec, req)

	// Проверяем результат - должен быть 404, так как chi не сопоставит маршрут
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestServeVideoFile_RangeRequest(t *testing.T) {
	// Создаем временную директорию для тестовых видео
	tmpDir := t.TempDir()
	testVideoPath := filepath.Join(tmpDir, "test.mp4")

	// Создаем тестовый файл с достаточным содержимым для проверки запроса диапазона
	testVideo := make([]byte, 1024) // 1KB видео
	for i := range testVideo {
		testVideo[i] = byte(i % 256) // Заполняем какими-то данными
	}

	err := os.WriteFile(testVideoPath, testVideo, 0644)
	assert.NoError(t, err)

	// Настраиваем конфигурацию
	cfg := &config.Config{
		Media: config.Media{
			VideoPatch: tmpDir,
		},
	}

	logger := logdiscart.NewDiscardLogger()

	// Создаем роутер с нужным параметром
	r := chi.NewRouter()
	r.Get("/video/{filename}", ServeVideoFile(cfg, logger))

	// Создаем запрос с заголовком Range
	req := httptest.NewRequest(http.MethodGet, "/video/test.mp4", nil)
	req.Header.Set("Range", "bytes=0-499") // Запрашиваем первые 500 байт
	rec := httptest.NewRecorder()

	// Выполняем запрос
	r.ServeHTTP(rec, req)

	// Проверяем результат
	assert.Equal(t, http.StatusPartialContent, rec.Code)
	assert.Equal(t, "bytes 0-499/1024", rec.Header().Get("Content-Range"))
	assert.Equal(t, "500", rec.Header().Get("Content-Length"))
	assert.Equal(t, testVideo[:500], rec.Body.Bytes())
}

func TestServeVideoFile_IfNoneMatchHeader(t *testing.T) {
	// Создаем временную директорию для тестовых видео
	tmpDir := t.TempDir()
	testVideoPath := filepath.Join(tmpDir, "test.mp4")

	// Создаем тестовый файл
	testVideo := []byte("fake video content")
	err := os.WriteFile(testVideoPath, testVideo, 0644)
	assert.NoError(t, err)

	// Настраиваем конфигурацию
	cfg := &config.Config{
		Media: config.Media{
			VideoPatch: tmpDir,
		},
	}

	logger := logdiscart.NewDiscardLogger()

	// Создаем роутер с нужным параметром
	r := chi.NewRouter()
	r.Get("/video/{filename}", ServeVideoFile(cfg, logger))

	// Сначала отправляем обычный запрос, чтобы получить ETag
	req := httptest.NewRequest(http.MethodGet, "/video/test.mp4", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Проверяем что ETag установлен
	actualEtag := rec.Header().Get("ETag")
	assert.NotEmpty(t, actualEtag)

	// Теперь отправляем запрос с If-None-Match равным ETag
	req = httptest.NewRequest(http.MethodGet, "/video/test.mp4", nil)
	req.Header.Set("If-None-Match", actualEtag)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Проверяем, что получили 304 Not Modified
	assert.Equal(t, http.StatusNotModified, rec.Code)
	assert.Empty(t, rec.Body.Bytes()) // Тело должно быть пустым
}

func TestGetStatusFromResponse(t *testing.T) {
	// Создаем ResponseWriter с фиксированным StatusCode
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusOK)

	// Проверяем функцию getStatusFromResponse
	status := getStatusFromResponse(rec)
	assert.Equal(t, http.StatusOK, status)

	// httptest.ResponseRecorder НЕ реализует интерфейс с методом Status()
	// поэтому getStatusFromResponse всегда вернет http.StatusOK для такого ResponseWriter
	// независимо от того, какой статус мы установили
	rec = httptest.NewRecorder()
	rec.WriteHeader(http.StatusNotFound)
	status = getStatusFromResponse(rec)
	// Ожидаем http.StatusOK, а НЕ http.StatusNotFound, потому что
	// getStatusFromResponse не может извлечь сохраненный статус
	assert.Equal(t, http.StatusOK, status)

	// Для полноценного тестирования нужно создать мок ResponseWriter,
	// который реализует метод Status()
	mockResponseWriter := &mockStatusResponseWriter{status: http.StatusNotFound}
	status = getStatusFromResponse(mockResponseWriter)
	assert.Equal(t, http.StatusNotFound, status)
}

// Создаем мок-объект для тестирования getStatusFromResponse
type mockStatusResponseWriter struct {
	status int
}

func (m *mockStatusResponseWriter) Header() http.Header {
	return http.Header{}
}

func (m *mockStatusResponseWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (m *mockStatusResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

// Метод Status, который ожидает getStatusFromResponse
func (m *mockStatusResponseWriter) Status() int {
	return m.status
}
