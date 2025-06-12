package v1

import (
	"database/sql"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetVideo_EmptyID(t *testing.T) {
	// Создаем моки и хендлер с помощью общей функции
	h, _, _ := newTestHandlerWithMocks(t)

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetVideo_FromCache(t *testing.T) {
	// Создаем моки и хендлер
	h, _, redisMock := newTestHandlerWithMocks(t)

	// Создаем данные видео и настраиваем мок Redis
	vid := &response.VideoResponse{ID: 1, Name: "video1"}
	vidData, err := json.Marshal(vid)
	require.NoError(t, err)

	redisMock.ExpectGet("video:1").SetVal(string(vidData))

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=1", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var got response.VideoResponse
	err = json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.Equal(t, vid.ID, got.ID)
	assert.Equal(t, vid.Name, got.Name)
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestGetVideo_FromStorageAndCache(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для отсутствия данных
	redisMock.ExpectGet("video:2").RedisNil()

	// Настраиваем мок SQL для получения видео
	// Используем актуальный SQL-запрос, соответствующий реализации в хранилище
	sqlMock.ExpectQuery(`SELECT v\.id, v\.url, v\.name, v\.description, c\.name as category, v\.img_url
        FROM videos v
        JOIN video_categories vc ON v\.id = vc\.video_id
        JOIN categories c ON vc\.category_id = c\.id
        WHERE v\.id = \$1 AND v\.deleted IS NOT TRUE`).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url", "name", "description", "category", "img_url"}).
			AddRow(2, "url2", "video2", "desc2", "category2", "img2"))

	// Важно: НЕ настраиваем ожидание вызова SetVideo, так как в реализации это делается асинхронно через горутину

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=2", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)

	// Выводим код статуса и тело ответа для отладки
	t.Logf("Status Code: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)

	var got response.VideoResponse
	err := json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.Equal(t, int64(2), got.ID)
	assert.Equal(t, "video2", got.Name)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	// Не проверяем ожидания Redis, так как вызов Set происходит асинхронно
}

func TestGetVideo_RedisError(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для ошибки
	redisMock.ExpectGet("video:3").SetErr(sql.ErrConnDone)

	// Настраиваем мок SQL для получения видео
	// Используем актуальный SQL-запрос, соответствующий реализации в хранилище
	sqlMock.ExpectQuery(`SELECT v\.id, v\.url, v\.name, v\.description, c\.name as category, v\.img_url
        FROM videos v
        JOIN video_categories vc ON v\.id = vc\.video_id
        JOIN categories c ON vc\.category_id = c\.id
        WHERE v\.id = \$1 AND v\.deleted IS NOT TRUE`).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url", "name", "description", "category", "img_url"}).
			AddRow(3, "url3", "video3", "desc3", "category3", "img3"))

	// Важно: НЕ настраиваем ожидание вызова SetVideo, так как в реализации это делается асинхронно через горутину

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=3", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)

	// Выводим код статуса и тело ответа для отладки
	t.Logf("Status Code: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)

	var got response.VideoResponse
	err := json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.Equal(t, int64(3), got.ID)
	assert.Equal(t, "video3", got.Name)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestGetVideo_StorageError(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для отсутствия данных
	redisMock.ExpectGet("video:4").RedisNil()

	// Настраиваем мок SQL для ошибки
	// Используем актуальный SQL-запрос, соответствующий реализации в хранилище
	sqlMock.ExpectQuery(`SELECT v\.id, v\.url, v\.name, v\.description, c\.name as category, v\.img_url
        FROM videos v
        JOIN video_categories vc ON v\.id = vc\.video_id
        JOIN categories c ON vc\.category_id = c\.id
        WHERE v\.id = \$1 AND v\.deleted IS NOT TRUE`).
		WithArgs(int64(4)).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=4", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)

	// Выводим код статуса и тело ответа для отладки
	t.Logf("Status Code: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestGetVideo_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для отсутствия данных
	redisMock.ExpectGet("video:999").RedisNil()

	// Настраиваем мок SQL для отсутствия результата
	// Используем актуальный SQL-запрос, соответствующий реализации в хранилище
	sqlMock.ExpectQuery(`SELECT v\.id, v\.url, v\.name, v\.description, c\.name as category, v\.img_url
        FROM videos v
        JOIN video_categories vc ON v\.id = vc\.video_id
        JOIN categories c ON vc\.category_id = c\.id
        WHERE v\.id = \$1 AND v\.deleted IS NOT TRUE`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=999", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)

	// Выводим код статуса и тело ответа для отладки
	t.Logf("Status Code: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestGetVideosByCategoryAndType_EmptyParams(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestHandlerWithMocks(t)

	// Проверка на пустой тип
	req1 := httptest.NewRequest(http.MethodGet, "/videos?type=&category=1", nil)
	w1 := httptest.NewRecorder()
	h.getVideosByCategoryAndType(w1, req1)
	assert.Equal(t, http.StatusBadRequest, w1.Code)

	// Проверка на пустую категорию
	req2 := httptest.NewRequest(http.MethodGet, "/videos?type=1&category=", nil)
	w2 := httptest.NewRecorder()
	h.getVideosByCategoryAndType(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestGetVideosByCategoryAndType_FromCache(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Создаем данные видео и настраиваем мок Redis
	videos := []response.VideoResponse{{ID: 1, Name: "video1", URL: "/video/url1"}, {ID: 2, Name: "video2", URL: "/video/url2"}}
	videosData, err := json.Marshal(videos)
	require.NoError(t, err)

	// Ожидания на проверки существования типа и категории (они всегда вызываются)
	sqlMock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM content_types WHERE id = \\$1 AND deleted IS NOT TRUE\\)").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	sqlMock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM categories WHERE id = \\$1 AND deleted IS NOT TRUE\\)").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Настраиваем возврат кэшированных данных
	redisMock.ExpectGet("videos:type:1:category:1").SetVal(string(videosData))

	req := httptest.NewRequest(http.MethodGet, "/videos?type=1&category=1", nil)
	w := httptest.NewRecorder()

	h.getVideosByCategoryAndType(w, req)

	t.Logf("Status Code: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)

	var got []response.VideoResponse
	err = json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err, "response should be valid JSON")
	assert.Len(t, got, 2)
	assert.Equal(t, videos[0].ID, got[0].ID)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestGetVideosByCategoryAndType_FromStorageAndCache(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для отсутствия данных
	redisMock.ExpectGet("videos:type:2:category:2").RedisNil()

	// Добавляем проверку существования типа контента
	sqlMock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM content_types WHERE id = \\$1 AND deleted IS NOT TRUE\\)").
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Добавляем проверку существования категории
	sqlMock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM categories WHERE id = \\$1 AND deleted IS NOT TRUE\\)").
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Настраиваем мок SQL для получения видео
	sqlMock.ExpectQuery("SELECT v.id, v.name, v.url, v.description, v.img_url, v.duration, v.service_name, v.created_at FROM videos v JOIN video_categories vc ON v.id = vc.video_id WHERE vc.category_id = \\$1 AND v.deleted IS NOT TRUE").
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "description", "img_url", "duration", "service_name", "created_at"}).
			AddRow(3, "video3", "/video/url3", "desc3", "/img/img3", 150, "youtube", time.Now()).
			AddRow(4, "video4", "/video/url4", "desc4", "/img/img4", 200, "vimeo", time.Now()))

	// Настраиваем мок Redis для сохранения в кэш
	redisMock.ExpectSet("videos:type:2:category:2", mock.Anything, time.Hour).SetVal("OK")

	req := httptest.NewRequest(http.MethodGet, "/videos?type=2&category=2", nil)
	w := httptest.NewRecorder()

	h.getVideosByCategoryAndType(w, req)

	// Выводим код статуса и тело ответа для отладки
	t.Logf("Status Code: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)

	var got []response.VideoResponse
	err := json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, int64(3), got[0].ID)
	assert.Equal(t, "video3", got[0].Name)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestGetVideosByCategoryAndType_StorageError(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для отсутствия данных
	redisMock.ExpectGet("videos:type:3:category:3").RedisNil()

	// Добавляем проверку существования типа контента
	sqlMock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM content_types WHERE id = \\$1 AND deleted IS NOT TRUE\\)").
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Настраиваем мок SQL для ошибки
	sqlMock.ExpectQuery("SELECT v.id, v.name, v.url, v.description, v.img_url, v.duration, v.service_name, v.created_at FROM videos v JOIN video_categories vc ON v.id = vc.video_id WHERE vc.category_id = \\$1 AND v.deleted IS NOT TRUE").
		WithArgs(int64(3)).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest(http.MethodGet, "/videos?type=3&category=3", nil)
	w := httptest.NewRecorder()

	h.getVideosByCategoryAndType(w, req)

	// Выводим код статуса и тело ответа для отладки
	t.Logf("Status Code: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}
