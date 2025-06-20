package docs

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterRoutes_SwaggerJSON(t *testing.T) {
	// Настройка тестового окружения
	r := chi.NewRouter()
	cfg := Config{User: "test", Password: "test"}

	// Создаём временную директорию и файл swagger.json для теста
	tempDir := t.TempDir()
	docsDir := filepath.Join(tempDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	swaggerJSON := `{"swagger": "2.0", "info": {"title": "Test API"}}`
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "swagger.json"), []byte(swaggerJSON), 0644))

	// Подменяем функцию getProjectRoot для теста
	originalGetProjectRoot := getProjectRoot
	defer func() { getProjectRoot = originalGetProjectRoot }()
	getProjectRoot = func() string { return tempDir }

	// Регистрируем маршруты и выполняем запрос
	RegisterRoutes(r, cfg)

	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Проверка результатов
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, swaggerJSON, rec.Body.String())
}

func TestRegisterRoutes_RapiDocHTML(t *testing.T) {
	// Настройка тестового окружения
	r := chi.NewRouter()
	cfg := Config{User: "test", Password: "test"}

	// Создаём временную директорию и файл rapidoc.html для теста
	tempDir := t.TempDir()
	docsDir := filepath.Join(tempDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	rapidocHTML := `<!DOCTYPE html><html><body><h1>RapiDoc Test</h1></body></html>`
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "rapidoc.html"), []byte(rapidocHTML), 0644))

	// Подменяем функцию getProjectRoot для теста
	originalGetProjectRoot := getProjectRoot
	defer func() { getProjectRoot = originalGetProjectRoot }()
	getProjectRoot = func() string { return tempDir }

	// Регистрируем маршруты и выполняем запрос
	RegisterRoutes(r, cfg)

	req := httptest.NewRequest(http.MethodGet, "/docs/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Проверка результатов
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "RapiDoc Test")
}

func TestRegisterRoutes_FileNotFound(t *testing.T) {
	// Настройка тестового окружения
	r := chi.NewRouter()
	cfg := Config{User: "test", Password: "test"}

	// Создаём пустую временную директорию (без нужных файлов)
	tempDir := t.TempDir()
	docsDir := filepath.Join(tempDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	// Подменяем функцию getProjectRoot для теста
	originalGetProjectRoot := getProjectRoot
	defer func() { getProjectRoot = originalGetProjectRoot }()
	getProjectRoot = func() string { return tempDir }

	// Регистрируем маршруты и выполняем запрос
	RegisterRoutes(r, cfg)

	// Тест на отсутствие swagger.json
	swaggerReq := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	swaggerRec := httptest.NewRecorder()
	r.ServeHTTP(swaggerRec, swaggerReq)
	assert.Equal(t, http.StatusNotFound, swaggerRec.Code)

	// Тест на отсутствие rapidoc.html
	docsReq := httptest.NewRequest(http.MethodGet, "/docs/", nil)
	docsRec := httptest.NewRecorder()
	r.ServeHTTP(docsRec, docsReq)
	assert.Equal(t, http.StatusNotFound, docsRec.Code)
}

func TestRegisterRoutes_StaticFiles(t *testing.T) {
	// Настройка тестового окружения
	r := chi.NewRouter()
	cfg := Config{User: "test", Password: "test"}

	// Создаём временную директорию и файл style.css для теста
	tempDir := t.TempDir()
	docsDir := filepath.Join(tempDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	cssContent := `body { color: red; }`
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "style.css"), []byte(cssContent), 0644))

	// Подменяем функцию getProjectRoot для теста
	originalGetProjectRoot := getProjectRoot
	defer func() { getProjectRoot = originalGetProjectRoot }()
	getProjectRoot = func() string { return tempDir }

	// Регистрируем маршруты и выполняем запрос
	RegisterRoutes(r, cfg)

	req := httptest.NewRequest(http.MethodGet, "/docs/style.css", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Проверка результатов
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "body { color: red; }")
}
