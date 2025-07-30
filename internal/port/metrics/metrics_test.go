package metrics_test

import (
	"context"
	"fmt"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/middleware/metrics"
	projectMetrics "github.com/langowen/bodybalance-backend/internal/port/metrics"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDataSourceMetrics(t *testing.T) {
	// Сбрасываем все метрики перед тестом
	projectMetrics.DataSourceRequests.Reset()

	// Создаем тестовый HTTP запрос
	req, err := http.NewRequest("GET", "/api/v1/video?video_id=123", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Тест для Redis источника данных
	t.Run("Redis Data Source", func(t *testing.T) {
		metrics.RecordDataSource(req, metrics.SourceRedis)

		// Проверяем, что метрика для Redis увеличилась
		count, err := getSingleMetricValue(projectMetrics.DataSourceRequests, map[string]string{
			"method":   "GET",
			"endpoint": "/api/v1/video",
			"source":   metrics.SourceRedis,
		})
		assert.NoError(t, err)
		assert.Equal(t, float64(1), count, "Redis metric should be incremented")
	})

	// Тест для SQL источника данных
	t.Run("SQL Data Source", func(t *testing.T) {
		metrics.RecordDataSource(req, metrics.SourceSQL)

		// Проверяем, что метрика для SQL увеличилась
		count, err := getSingleMetricValue(projectMetrics.DataSourceRequests, map[string]string{
			"method":   "GET",
			"endpoint": "/api/v1/video",
			"source":   metrics.SourceSQL,
		})
		assert.NoError(t, err)
		assert.Equal(t, float64(1), count, "SQL metric should be incremented")

		// Проверяем, что метрика для Redis не изменилась
		redisCount, err := getSingleMetricValue(projectMetrics.DataSourceRequests, map[string]string{
			"method":   "GET",
			"endpoint": "/api/v1/video",
			"source":   metrics.SourceRedis,
		})
		assert.NoError(t, err)
		assert.Equal(t, float64(1), redisCount, "Redis metric should not change")
	})
}

func TestContextFunctions(t *testing.T) {
	// Тест для WithDataSource и GetDataSource
	t.Run("Context Data Source", func(t *testing.T) {
		ctx := context.Background()

		// Добавляем информацию о Redis источнике в контекст
		ctxWithRedis := metrics.WithDataSource(ctx, metrics.SourceRedis)
		source, ok := metrics.GetDataSource(ctxWithRedis)
		assert.True(t, ok, "Should find data source in context")
		assert.Equal(t, metrics.SourceRedis, source, "Source should be Redis")

		// Добавляем информацию о SQL источнике в контекст
		ctxWithSQL := metrics.WithDataSource(ctx, metrics.SourceSQL)
		source, ok = metrics.GetDataSource(ctxWithSQL)
		assert.True(t, ok, "Should find data source in context")
		assert.Equal(t, metrics.SourceSQL, source, "Source should be SQL")

		// Проверяем отсутствие информации в исходном контексте
		source, ok = metrics.GetDataSource(ctx)
		assert.False(t, ok, "Should not find data source in original context")
		assert.Empty(t, source, "Source should be empty")
	})
}

func TestDataSourceMiddleware(t *testing.T) {
	// Сбрасываем все метрики перед тестом
	projectMetrics.DataSourceRequests.Reset()

	// Создаем тестовый обработчик
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Имитируем обработчик API, который записывает метрики

		// Симулируем случай, когда данные из Redis
		if r.URL.Path == "/with-redis" {
			metrics.RecordDataSource(r, metrics.SourceRedis)
		} else {
			// Симулируем случай, когда данные из SQL
			metrics.RecordDataSource(r, metrics.SourceSQL)
		}

		w.WriteHeader(http.StatusOK)
	})

	// Тест для запроса с данными из Redis
	t.Run("Request with Redis source", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/with-redis", nil)
		rr := httptest.NewRecorder()

		testHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Проверяем метрики
		count, err := getSingleMetricValue(projectMetrics.DataSourceRequests, map[string]string{
			"method":   "GET",
			"endpoint": "/with-redis",
			"source":   metrics.SourceRedis,
		})
		assert.NoError(t, err)
		assert.Equal(t, float64(1), count, "Redis metric should be incremented")
	})

	// Тест для запроса с данными из SQL
	t.Run("Request with SQL source", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/without-redis", nil)
		rr := httptest.NewRecorder()

		testHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Проверяем метрики
		count, err := getSingleMetricValue(projectMetrics.DataSourceRequests, map[string]string{
			"method":   "GET",
			"endpoint": "/without-redis",
			"source":   metrics.SourceSQL,
		})
		assert.NoError(t, err)
		assert.Equal(t, float64(1), count, "SQL metric should be incremented")
	})
}

func TestAPIRequestsTotal(t *testing.T) {
	// Сбрасываем метрики перед тестом
	projectMetrics.APIRequestsTotal.Reset()

	// Проверяем счетчик запросов к API
	t.Run("API requests counter", func(t *testing.T) {
		// Увеличиваем счетчик с разными лейблами
		projectMetrics.APIRequestsTotal.WithLabelValues("GET", "/api/v1/video", "200").Inc()
		projectMetrics.APIRequestsTotal.WithLabelValues("GET", "/api/v1/video", "404").Inc()
		projectMetrics.APIRequestsTotal.WithLabelValues("POST", "/api/v1/login", "200").Inc()

		// Проверяем, что метрики увеличились
		count1 := testutil.ToFloat64(projectMetrics.APIRequestsTotal.WithLabelValues("GET", "/api/v1/video", "200"))
		assert.Equal(t, float64(1), count1, "API GET /video 200 metric should be 1")

		count2 := testutil.ToFloat64(projectMetrics.APIRequestsTotal.WithLabelValues("GET", "/api/v1/video", "404"))
		assert.Equal(t, float64(1), count2, "API GET /video 404 metric should be 1")

		count3 := testutil.ToFloat64(projectMetrics.APIRequestsTotal.WithLabelValues("POST", "/api/v1/login", "200"))
		assert.Equal(t, float64(1), count3, "API POST /login 200 metric should be 1")

		// Увеличиваем еще раз и проверяем
		projectMetrics.APIRequestsTotal.WithLabelValues("GET", "/api/v1/video", "200").Inc()
		count4 := testutil.ToFloat64(projectMetrics.APIRequestsTotal.WithLabelValues("GET", "/api/v1/video", "200"))
		assert.Equal(t, float64(2), count4, "API GET /video 200 metric should be 2 after increment")
	})
}

func TestAdminRequestsTotal(t *testing.T) {
	// Сбрасываем метрики перед тестом
	projectMetrics.AdminRequestsTotal.Reset()

	// Проверяем счетчик запросов к админке
	t.Run("Admin requests counter", func(t *testing.T) {
		// Увеличиваем счетчик с разными лейблами
		projectMetrics.AdminRequestsTotal.WithLabelValues("GET", "/admin/category", "200").Inc()
		projectMetrics.AdminRequestsTotal.WithLabelValues("PUT", "/admin/category/1", "200").Inc()
		projectMetrics.AdminRequestsTotal.WithLabelValues("DELETE", "/admin/video/1", "404").Inc()

		// Проверяем, что метрики увеличились
		count1 := testutil.ToFloat64(projectMetrics.AdminRequestsTotal.WithLabelValues("GET", "/admin/category", "200"))
		assert.Equal(t, float64(1), count1, "Admin GET /category metric should be 1")

		count2 := testutil.ToFloat64(projectMetrics.AdminRequestsTotal.WithLabelValues("PUT", "/admin/category/1", "200"))
		assert.Equal(t, float64(1), count2, "Admin PUT /category/1 metric should be 1")

		count3 := testutil.ToFloat64(projectMetrics.AdminRequestsTotal.WithLabelValues("DELETE", "/admin/video/1", "404"))
		assert.Equal(t, float64(1), count3, "Admin DELETE /video/1 404 metric should be 1")
	})
}

func TestRequestDuration(t *testing.T) {
	// Сбрасываем метрики перед тестом
	projectMetrics.RequestDuration.Reset()

	// Проверяем гистограмму длительности запросов
	t.Run("Request duration histogram", func(t *testing.T) {
		// Наблюдаем разные значения длительности
		projectMetrics.RequestDuration.WithLabelValues("api", "GET", "/api/v1/video").Observe(0.1)
		projectMetrics.RequestDuration.WithLabelValues("api", "GET", "/api/v1/video").Observe(0.2)
		projectMetrics.RequestDuration.WithLabelValues("admin", "POST", "/admin/category").Observe(0.5)

		// Сложно напрямую проверить значения гистограммы, поэтому используем метод ToString
		// для отладки и мягкой проверки наличия метрики
		metricStr := testutil.CollectAndCount(projectMetrics.RequestDuration)

		// Проверяем, что метрики были собраны
		assert.Greater(t, metricStr, 0, "Histogram should have metrics")
	})
}

func TestActiveRequests(t *testing.T) {
	// Сбрасываем метрики перед тестом
	projectMetrics.ActiveRequests.Reset()

	// Проверяем gauge активных запросов
	t.Run("Active requests gauge", func(t *testing.T) {
		// Увеличиваем и уменьшаем значение gauge
		projectMetrics.ActiveRequests.WithLabelValues("api").Inc()
		projectMetrics.ActiveRequests.WithLabelValues("api").Inc()
		projectMetrics.ActiveRequests.WithLabelValues("admin").Inc()

		// Проверяем значения
		gauge1 := testutil.ToFloat64(projectMetrics.ActiveRequests.WithLabelValues("api"))
		assert.Equal(t, float64(2), gauge1, "Active API requests gauge should be 2")

		gauge2 := testutil.ToFloat64(projectMetrics.ActiveRequests.WithLabelValues("admin"))
		assert.Equal(t, float64(1), gauge2, "Active admin requests gauge should be 1")

		// Уменьшаем значение и проверяем
		projectMetrics.ActiveRequests.WithLabelValues("api").Dec()
		gauge3 := testutil.ToFloat64(projectMetrics.ActiveRequests.WithLabelValues("api"))
		assert.Equal(t, float64(1), gauge3, "Active API requests gauge should be 1 after decrement")
	})
}

func TestStaticFileRequests(t *testing.T) {
	// Сбрасываем метрики перед тестом
	projectMetrics.StaticFileRequests.Reset()

	// Проверяем счетчик запросов к статическим файлам
	t.Run("Static file requests counter", func(t *testing.T) {
		// Увеличиваем счетчик с разными лейблами
		projectMetrics.StaticFileRequests.WithLabelValues("image", "/img/placeholder.jpg", "200").Inc()
		projectMetrics.StaticFileRequests.WithLabelValues("video", "/video/sample-5s.mp4", "200").Inc()
		projectMetrics.StaticFileRequests.WithLabelValues("image", "/img/nonexistent.jpg", "404").Inc()

		// Проверяем, что метрики увеличились
		count1 := testutil.ToFloat64(projectMetrics.StaticFileRequests.WithLabelValues("image", "/img/placeholder.jpg", "200"))
		assert.Equal(t, float64(1), count1, "Static image request metric should be 1")

		count2 := testutil.ToFloat64(projectMetrics.StaticFileRequests.WithLabelValues("video", "/video/sample-5s.mp4", "200"))
		assert.Equal(t, float64(1), count2, "Static video request metric should be 1")

		count3 := testutil.ToFloat64(projectMetrics.StaticFileRequests.WithLabelValues("image", "/img/nonexistent.jpg", "404"))
		assert.Equal(t, float64(1), count3, "Static 404 image request metric should be 1")
	})
}

func TestStaticFileSize(t *testing.T) {
	// Сбрасываем метрики перед тестом
	projectMetrics.StaticFileSize.Reset()

	// Проверяем гистограмму размера статических файлов
	t.Run("Static file size histogram", func(t *testing.T) {
		// Наблюдаем разные значения размеров
		projectMetrics.StaticFileSize.WithLabelValues("image", "placeholder.jpg").Observe(1024 * 50)     // 50KB
		projectMetrics.StaticFileSize.WithLabelValues("video", "sample-5s.mp4").Observe(1024 * 1024 * 2) // 2MB

		// Используем CollectAndCount для проверки наличия метрик вместо ToFloat64
		metricCount := testutil.CollectAndCount(projectMetrics.StaticFileSize)

		// Проверяем, что метрики были собраны
		assert.Greater(t, metricCount, 0, "Histogram should have metrics")
	})
}

// Вспомогательная функция для получения значения метрики с заданными лейблами
func getSingleMetricValue(metric *prometheus.CounterVec, labels map[string]string) (float64, error) {
	// Метод WithLabelValues принимает значения лейблов в определенном порядке
	// который должен соответствовать порядку лейблов при создании CounterVec

	// Для DataSourceRequests порядок: method, endpoint, source
	var labelValues []string

	// Определяем, какая метрика нам нужна и формируем соответствующий массив
	if metric == projectMetrics.DataSourceRequests {
		labelValues = []string{
			labels["method"],
			labels["endpoint"],
			labels["source"],
		}
	} else {
		// Для других метрик нужно будет расширить этот код
		return 0, fmt.Errorf("unsupported metric type")
	}

	// Получаем значение метрики с указанными лейблами
	value := testutil.ToFloat64(metric.WithLabelValues(labelValues...))
	return value, nil
}
