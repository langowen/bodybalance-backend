// filepath: /Users/sergejmedvedev/Library/CloudStorage/SynologyDrive-homelang/dev/project/bodybalance-backend/internal/http-server/middleware/metrics/datasource_test.go
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

	// Тест для redis источника данных
	t.Run("redis Data Source", func(t *testing.T) {
		metrics.RecordDataSource(req, metrics.SourceRedis)

		// Проверяем, что метрика для redis увеличилась
		count, err := getSingleMetricValue(projectMetrics.DataSourceRequests, map[string]string{
			"method":   "GET",
			"endpoint": "/api/v1/video",
			"source":   metrics.SourceRedis,
		})
		assert.NoError(t, err)
		assert.Equal(t, float64(1), count, "redis metric should be incremented")
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

		// Проверяем, что метрика для redis не изменилась
		redisCount, err := getSingleMetricValue(projectMetrics.DataSourceRequests, map[string]string{
			"method":   "GET",
			"endpoint": "/api/v1/video",
			"source":   metrics.SourceRedis,
		})
		assert.NoError(t, err)
		assert.Equal(t, float64(1), redisCount, "redis metric should not change")
	})
}

func TestContextFunctions(t *testing.T) {
	// Тест для WithDataSource и GetDataSource
	t.Run("Context Data Source", func(t *testing.T) {
		ctx := context.Background()

		// Добавляем информацию о redis источнике в контекст
		ctxWithRedis := metrics.WithDataSource(ctx, metrics.SourceRedis)
		source, ok := metrics.GetDataSource(ctxWithRedis)
		assert.True(t, ok, "Should find data source in context")
		assert.Equal(t, metrics.SourceRedis, source, "Source should be redis")

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

		// Симулируем случай, когда данные из redis
		if r.URL.Path == "/with-redis" {
			metrics.RecordDataSource(r, metrics.SourceRedis)
		} else {
			// Симулируем случай, когда данные из SQL
			metrics.RecordDataSource(r, metrics.SourceSQL)
		}

		w.WriteHeader(http.StatusOK)
	})

	// Тест для запроса с данными из redis
	t.Run("Request with redis source", func(t *testing.T) {
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
		assert.Equal(t, float64(1), count, "redis metric should be incremented")
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
