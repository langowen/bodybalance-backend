package admin

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тест на успешное добавление нового видео
func TestHandler_AddVideo_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД для добавления видео
	sqlMock.ExpectQuery(`INSERT INTO videos \(url, name, description, img_url, deleted\) VALUES \(\$1, \$2, \$3, \$4, FALSE\) RETURNING id`).
		WithArgs("http://example.com/test.mp4", "Test Video", "Test Description", "http://example.com/test.jpg").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Настраиваем ожидания для транзакции
	sqlMock.ExpectBegin()
	sqlMock.ExpectPrepare(`INSERT INTO video_categories \(video_id, category_id\) VALUES \(\$1, \$2\) ON CONFLICT \(video_id, category_id\) DO NOTHING`)
	sqlMock.ExpectExec(`INSERT INTO video_categories \(video_id, category_id\) VALUES \(\$1, \$2\) ON CONFLICT \(video_id, category_id\) DO NOTHING`).
		WithArgs(int64(1), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	sqlMock.ExpectCommit()

	// Создаем тестовый запрос
	req := admResponse.VideoRequest{
		Name:        "Test Video",
		URL:         "http://example.com/test.mp4",
		Description: "Test Description",
		ImgURL:      "http://example.com/test.jpg",
		CategoryIDs: []int64{1},
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/admin/video", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addVideo(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Проверяем ответ
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(1), response["id"])
	assert.Contains(t, response["message"], "successfully")
}

// Тест на ошибку валидации при создании видео
func TestHandler_AddVideo_ValidationError(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем тестовый запрос с пустыми обязательными полями
	req := admResponse.VideoRequest{
		Name:        "", // Пустое имя
		URL:         "", // Пустой URL
		Description: "Test Description",
		ImgURL:      "http://example.com/test.jpg",
		CategoryIDs: []int64{1},
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/admin/video", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addVideo(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на невалидный формат JSON
func TestHandler_AddVideo_InvalidJSON(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем тестовый запрос с невалидным JSON
	httpReq := httptest.NewRequest(http.MethodPost, "/admin/video", bytes.NewBuffer([]byte(`{"name": "Test Video", "url": `)))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addVideo(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на успешное получение видео по ID
func TestHandler_GetVideo_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	timeNow := time.Now()

	// Настраиваем ожидание запроса к БД для получения видео
	// Исправляем запрос и возвращаемые поля в соответствии с реализацией GetVideo
	sqlMock.ExpectQuery(`SELECT id, url, name, description, img_url, created_at FROM videos WHERE id = \$1 AND deleted IS NOT TRUE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url", "name", "description", "img_url", "created_at"}).
			AddRow(1, "http://example.com/test.mp4", "Test Video", "Test Description", "http://example.com/test.jpg", timeNow))

	// Настраиваем ожидание запроса к БД для получения категорий
	sqlMock.ExpectQuery(`SELECT c.id, c.name FROM categories c JOIN video_categories vc ON c.id = vc.category_id WHERE vc.video_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Категория 1"))

	// Добавляем мок для запроса типов контента для категории
	sqlMock.ExpectQuery(`SELECT ct.id, ct.name FROM content_types ct JOIN category_content_types cct ON ct.id = cct.content_type_id WHERE cct.category_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Тип 1"))

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Get("/admin/video/{id}", h.getVideo)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/video/1", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Проверяем ответ
	var response admResponse.VideoResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, "Test Video", response.Name)
	assert.Equal(t, "http://example.com/test.mp4", response.URL)
	assert.Equal(t, timeNow.Format("02.01.2006"), response.DateCreated)
	assert.Equal(t, 1, len(response.Categories))
	assert.Equal(t, int64(1), response.Categories[0].ID)
	assert.Equal(t, 1, len(response.Categories[0].Types))
	assert.Equal(t, int64(1), response.Categories[0].Types[0].ID)
}

// Тест на получение несуществующего видео
func TestHandler_GetVideo_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД с ошибкой
	sqlMock.ExpectQuery(`SELECT id, url, name, description, img_url, created_at FROM videos WHERE id = \$1 AND deleted IS NOT TRUE`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Get("/admin/video/{id}", h.getVideo)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/video/999", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на невалидный ID видео
func TestHandler_GetVideo_InvalidID(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Get("/admin/video/{id}", h.getVideo)

	// Создаем тестовый запрос с невалидным ID
	req := httptest.NewRequest(http.MethodGet, "/admin/video/invalid", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на успешное получение всех видео
func TestHandler_GetVideos_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД для получения списка видео
	// Исправляем запрос в соответствии с реальной реализацией
	timeNow := time.Now()
	sqlMock.ExpectQuery(`SELECT id, url, name, description, img_url, created_at FROM videos WHERE deleted IS NOT TRUE ORDER BY id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url", "name", "description", "img_url", "created_at"}).
			AddRow(1, "http://example.com/1.mp4", "Video 1", "Description 1", "http://example.com/1.jpg", timeNow).
			AddRow(2, "http://example.com/2.mp4", "Video 2", "Description 2", "http://example.com/2.jpg", timeNow))

	// Настраиваем ожидания запросов к БД для получения категорий для каждого видео
	sqlMock.ExpectQuery(`SELECT c.id, c.name FROM categories c JOIN video_categories vc ON c.id = vc.category_id WHERE vc.video_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Категория 1"))

	// Добавляем мок для запроса типов контента для первого видео
	sqlMock.ExpectQuery(`SELECT ct.id, ct.name FROM content_types ct JOIN category_content_types cct ON ct.id = cct.content_type_id WHERE cct.category_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Тип 1"))

	sqlMock.ExpectQuery(`SELECT c.id, c.name FROM categories c JOIN video_categories vc ON c.id = vc.category_id WHERE vc.video_id = \$1`).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(2, "Категория 2"))

	// Добавляем мок для запроса типов контента для второго видео
	sqlMock.ExpectQuery(`SELECT ct.id, ct.name FROM content_types ct JOIN category_content_types cct ON ct.id = cct.content_type_id WHERE cct.category_id = \$1`).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(2, "Тип 2"))

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/video", nil)
	w := httptest.NewRecorder()

	// Вызываем метод
	h.getVideos(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Проверяем ответ
	var response []admResponse.VideoResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Video 1", response[0].Name)
	assert.Equal(t, "Video 2", response[1].Name)
	assert.Equal(t, int64(1), response[0].ID)
	assert.Equal(t, int64(2), response[1].ID)
}

// Тест на успешное обновление видео
func TestHandler_UpdateVideo_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание для удаления из Redis
	redisMock.ExpectDel("video:1").SetVal(1)

	// Настраиваем ожидание запроса к БД для обновления видео
	// Исправляем порядок параметров, чтобы он соответствовал реальной реализации
	sqlMock.ExpectExec(`UPDATE videos SET url = \$1, name = \$2, description = \$3, img_url = \$4 WHERE id = \$5 AND deleted IS NOT TRUE`).
		WithArgs("http://example.com/updated.mp4", "Updated Video", "Updated Description", "http://example.com/updated.jpg", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Настраиваем ожидание запросов для обновления категорий
	sqlMock.ExpectExec(`DELETE FROM video_categories WHERE video_id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Используем Begin-Prepare-Exec-Commit для добавления категорий как в AddVideoCategories
	sqlMock.ExpectBegin()
	sqlMock.ExpectPrepare(`INSERT INTO video_categories \(video_id, category_id\) VALUES \(\$1, \$2\) ON CONFLICT \(video_id, category_id\) DO NOTHING`)
	sqlMock.ExpectExec(`INSERT INTO video_categories \(video_id, category_id\) VALUES \(\$1, \$2\) ON CONFLICT \(video_id, category_id\) DO NOTHING`).
		WithArgs(int64(1), int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	sqlMock.ExpectCommit()

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Put("/admin/video/{id}", h.updateVideo)

	// Создаем тестовый запрос
	req := admResponse.VideoRequest{
		Name:        "Updated Video",
		URL:         "http://example.com/updated.mp4",
		Description: "Updated Description",
		ImgURL:      "http://example.com/updated.jpg",
		CategoryIDs: []int64{2},
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPut, "/admin/video/1", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на обновление несуществующего видео
func TestHandler_UpdateVideo_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание для удаления из Redis
	redisMock.ExpectDel("video:999").SetVal(0)

	// Настраиваем ожидание запроса к БД с ошибкой
	// Исправляем порядок параметров
	sqlMock.ExpectExec(`UPDATE videos SET url = \$1, name = \$2, description = \$3, img_url = \$4 WHERE id = \$5 AND deleted IS NOT TRUE`).
		WithArgs("http://example.com/updated.mp4", "Updated Video", "Updated Description", "http://example.com/updated.jpg", int64(999)).
		WillReturnError(sql.ErrNoRows)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Put("/admin/video/{id}", h.updateVideo)

	// Создаем тестовый запрос
	req := admResponse.VideoRequest{
		Name:        "Updated Video",
		URL:         "http://example.com/updated.mp4",
		Description: "Updated Description",
		ImgURL:      "http://example.com/updated.jpg",
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPut, "/admin/video/999", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на успешное удаление видео
func TestHandler_DeleteVideo_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД для удаления видео
	sqlMock.ExpectExec(`UPDATE videos SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Delete("/admin/video/{id}", h.deleteVideo)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodDelete, "/admin/video/1", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Проверяем ответ
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(1), response["id"])
	assert.Contains(t, response["message"], "deleted successfully")
}

// Тест на удаление несуществующего видео
func TestHandler_DeleteVideo_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД с ошибкой
	sqlMock.ExpectExec(`UPDATE videos SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Delete("/admin/video/{id}", h.deleteVideo)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodDelete, "/admin/video/999", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на невалидный ID при удалении видео
func TestHandler_DeleteVideo_InvalidID(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Delete("/admin/video/{id}", h.deleteVideo)

	// Создаем тестовый запрос с невалидным ID
	req := httptest.NewRequest(http.MethodDelete, "/admin/video/invalid", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
