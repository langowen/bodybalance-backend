package v1

import (
	"database/sql"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCategoriesByType_EmptyType(t *testing.T) {
	// Создаем моки и хендлер с помощью общей функции
	h, _, _ := newTestHandlerWithMocks(t)

	req := httptest.NewRequest(http.MethodGet, "/category?type=", nil)
	w := httptest.NewRecorder()

	h.getCategoriesByType(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCategoriesByType_FromCache(t *testing.T) {
	// Создаем моки и хендлер
	h, _, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для получения категорий из кэша
	cats := []response.CategoryResponse{{ID: 1, Name: "cat1", ImgURL: "img1"}}
	catsData, err := json.Marshal(cats)
	require.NoError(t, err)

	redisMock.ExpectGet("categories:1").SetVal(string(catsData))

	req := httptest.NewRequest(http.MethodGet, "/category?type=1", nil)
	w := httptest.NewRecorder()

	h.getCategoriesByType(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var got []response.CategoryResponse
	err = json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.Equal(t, cats[0].ID, got[0].ID)
	assert.Equal(t, cats[0].Name, got[0].Name)
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestGetCategoriesByType_FromStorageAndCache(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для отсутствия данных
	redisMock.ExpectGet("categories:2").RedisNil()

	// Сначала настраиваем ожидание запроса для проверки существования типа контента (chekType)
	sqlMock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Настраиваем мок SQL для получения категорий
	// Используем актуальный SQL-запрос, соответствующий реализации в хранилище
	sqlMock.ExpectQuery(`SELECT c.id, c.name, c.img_url FROM categories c JOIN category_content_types cct ON c.id = cct.category_id JOIN content_types ct ON cct.content_type_id = ct.id WHERE ct.id = \$1 AND c.deleted IS NOT TRUE ORDER BY c.created_at DESC`).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "img_url"}).
			AddRow(1, "cat1", "img1").
			AddRow(2, "cat2", "img2"))

	// Важно: НЕ настраиваем ожидание вызова SetCategories, так как в реализации это делается асинхронно через горутину

	req := httptest.NewRequest(http.MethodGet, "/category?type=2", nil)
	w := httptest.NewRecorder()

	h.getCategoriesByType(w, req)

	// Выводим код статуса и тело ответа для отладки
	t.Logf("Status Code: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)

	var got []response.CategoryResponse
	err := json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, int64(1), got[0].ID)
	assert.Equal(t, "cat1", got[0].Name)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	// Не проверяем ожидания Redis, так как вызов Set происходит асинхронно
}

func TestGetCategoriesByType_RedisError(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для ошибки
	redisMock.ExpectGet("categories:3").SetErr(sql.ErrConnDone)

	// Сначала настраиваем ожидание запроса для проверки существования типа контента (chekType)
	sqlMock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Настраиваем мок SQL для получения категорий
	// Используем актуальный SQL-запрос, соответствующий реализации в хранилище
	sqlMock.ExpectQuery(`SELECT c.id, c.name, c.img_url FROM categories c JOIN category_content_types cct ON c.id = cct.category_id JOIN content_types ct ON cct.content_type_id = ct.id WHERE ct.id = \$1 AND c.deleted IS NOT TRUE ORDER BY c.created_at DESC`).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "img_url"}).
			AddRow(3, "cat3", "img3"))

	// Важно: НЕ настраиваем ожидание вызова SetCategories, так как в реализации это делается асинхронно через горутину

	req := httptest.NewRequest(http.MethodGet, "/category?type=3", nil)
	w := httptest.NewRecorder()

	h.getCategoriesByType(w, req)

	// Выводим код статуса и тело ответа для отладки
	t.Logf("Status Code: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)

	var got []response.CategoryResponse
	err := json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, int64(3), got[0].ID)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestGetCategoriesByType_StorageError(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для отсутствия данных
	redisMock.ExpectGet("categories:4").RedisNil()

	// Сначала настраиваем ожидание запроса для проверки существования типа контента (chekType)
	sqlMock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
		WithArgs(int64(4)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Настраиваем мок SQL для ошибки
	// Используем актуальный SQL-запрос, соответствующий реализации в хранилище
	sqlMock.ExpectQuery(`SELECT c.id, c.name, c.img_url FROM categories c JOIN category_content_types cct ON c.id = cct.category_id JOIN content_types ct ON cct.content_type_id = ct.id WHERE ct.id = \$1 AND c.deleted IS NOT TRUE ORDER BY c.created_at DESC`).
		WithArgs(int64(4)).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest(http.MethodGet, "/category?type=4", nil)
	w := httptest.NewRecorder()

	h.getCategoriesByType(w, req)

	// Выводим код статуса и тело ответа для отладки
	t.Logf("Status Code: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestGetCategoriesByType_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для отсутствия данных
	redisMock.ExpectGet("categories:5").RedisNil()

	// Особый случай: для проверки ситуации "не найдено" мы должны настроить chekType
	// на возврат false, чтобы показать отсутствие типа контента
	sqlMock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// В этом случае второй запрос не будет выполнен, так как chekType вернет ошибку

	req := httptest.NewRequest(http.MethodGet, "/category?type=5", nil)
	w := httptest.NewRecorder()

	h.getCategoriesByType(w, req)

	// Выводим код статуса и тело ответа для отладки
	t.Logf("Status Code: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}
