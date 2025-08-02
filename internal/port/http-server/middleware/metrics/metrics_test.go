package metrics_test

import (
	mw "github.com/langowen/bodybalance-backend/internal/port/http-server/middleware/metrics"
	projectMetrics "github.com/langowen/bodybalance-backend/internal/port/metrics"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestRequestMetricsMiddleware(t *testing.T) {
	// Тест для API метрик middleware
	t.Run("API metrics middleware", func(t *testing.T) {
		// Сбрасываем все метрики перед тестом
		projectMetrics.APIRequestsTotal.Reset()
		projectMetrics.RequestDuration.Reset()
		projectMetrics.ActiveRequests.Reset()

		// Создаем тестовый обработчик
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Имитируем задержку запроса
			time.Sleep(10 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test dto"))
		})

		// Применяем middleware
		apiMiddleware := mw.Middleware("api")
		wrappedHandler := apiMiddleware(handler)

		// Создаем тестовый запрос
		req := httptest.NewRequest("GET", "/api/v1/video", nil)
		rr := httptest.NewRecorder()

		// Создаем роутер chi и регистрируем маршрут для корректного определения пути
		router := chi.NewRouter()
		router.Get("/api/v1/video", func(w http.ResponseWriter, r *http.Request) {
			wrappedHandler.ServeHTTP(w, r)
		})
		router.ServeHTTP(rr, req)

		// Проверяем HTTP-ответ
		assert.Equal(t, http.StatusOK, rr.Code)

		// Проверяем метрики
		apiMetricCount := testutil.ToFloat64(projectMetrics.APIRequestsTotal.WithLabelValues("GET", "/api/v1/video", "200"))
		assert.Equal(t, float64(1), apiMetricCount, "API request total metric should be 1")

		// Проверяем, что ActiveRequests вернулся к 0 после запроса
		activeCount := testutil.ToFloat64(projectMetrics.ActiveRequests.WithLabelValues("api"))
		assert.Equal(t, float64(0), activeCount, "Active requests should be 0 after request completion")

		// Проверяем, что метрика длительности была создана
		durationMetricCount := testutil.CollectAndCount(projectMetrics.RequestDuration)
		assert.Greater(t, durationMetricCount, 0, "Request duration metric should be recorded")
	})

	// Тест для Admin метрик middleware
	t.Run("Admin metrics middleware", func(t *testing.T) {
		// Сбрасываем все метрики перед тестом
		projectMetrics.AdminRequestsTotal.Reset()
		projectMetrics.RequestDuration.Reset()
		projectMetrics.ActiveRequests.Reset()

		// Создаем тестовый обработчик
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		})

		// Применяем middleware
		adminMiddleware := mw.Middleware("admin")
		wrappedHandler := adminMiddleware(handler)

		// Создаем тестовый запрос
		req := httptest.NewRequest("POST", "/admin/category", nil)
		rr := httptest.NewRecorder()

		// Создаем роутер chi и регистрируем маршрут для корректного определения пути
		router := chi.NewRouter()
		router.Post("/admin/category", func(w http.ResponseWriter, r *http.Request) {
			wrappedHandler.ServeHTTP(w, r)
		})
		router.ServeHTTP(rr, req)

		// Проверяем HTTP-ответ
		assert.Equal(t, http.StatusOK, rr.Code)

		// Проверяем метрики
		adminMetricCount := testutil.ToFloat64(projectMetrics.AdminRequestsTotal.WithLabelValues("POST", "/admin/category", "200"))
		assert.Equal(t, float64(1), adminMetricCount, "Admin request total metric should be 1")

		// Проверяем, что ActiveRequests вернулся к 0 после запроса
		activeCount := testutil.ToFloat64(projectMetrics.ActiveRequests.WithLabelValues("admin"))
		assert.Equal(t, float64(0), activeCount, "Active requests should be 0 after request completion")

		// Проверяем, что метрика длительности была создана
		durationMetricCount := testutil.CollectAndCount(projectMetrics.RequestDuration)
		assert.Greater(t, durationMetricCount, 0, "Request duration metric should be recorded")
	})

	// Тест обработки ошибки
	t.Run("Error dto metrics", func(t *testing.T) {
		// Сбрасываем метрики перед тестом
		projectMetrics.APIRequestsTotal.Reset()

		// Создаем тестовый обработчик с ошибкой
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		// Применяем middleware
		apiMiddleware := mw.Middleware("api")
		wrappedHandler := apiMiddleware(handler)

		// Создаем тестовый запрос
		req := httptest.NewRequest("GET", "/api/v1/video/999", nil)
		rr := httptest.NewRecorder()

		// Создаем роутер chi и регистрируем маршрут
		router := chi.NewRouter()
		router.Get("/api/v1/video/{id}", func(w http.ResponseWriter, r *http.Request) {
			wrappedHandler.ServeHTTP(w, r)
		})
		router.ServeHTTP(rr, req)

		// Проверяем HTTP-ответ
		assert.Equal(t, http.StatusNotFound, rr.Code)

		// Проверяем метрики с кодом ошибки
		errorMetricCount := testutil.ToFloat64(projectMetrics.APIRequestsTotal.WithLabelValues("GET", "/api/v1/video/{id}", "404"))
		assert.Equal(t, float64(1), errorMetricCount, "API error request metric should be 1")
	})
}
