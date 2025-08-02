package admin

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin/dto"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тест на успешное добавление нового типа контента
func TestHandler_AddType_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`INSERT INTO content_types \(name, deleted\) VALUES \(\$1, FALSE\) RETURNING id, name, created_at`).
		WithArgs("Йога").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at"}).
			AddRow(1, "Йога", time.Now()))

	// Создаем тестовый запрос
	body := []byte(`{"name": "Йога"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/type", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addType(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Сначала проверяем фактический формат ответа
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Проверяем, что в ответе есть поле id и message
	assert.Contains(t, response, "id")
	assert.Contains(t, response, "message")

	// Проверяем значение message
	assert.Contains(t, response["message"], "successfully")

	// Проверяем, что id является объектом типа TypeResponse
	typeResponse, ok := response["id"].(map[string]interface{})
	require.True(t, ok, "id не является объектом")

	// Проверяем поля в объекте TypeResponse
	assert.Contains(t, typeResponse, "id")
	assert.Contains(t, typeResponse, "name")
	assert.Contains(t, typeResponse, "created_at")

	// Проверяем значения полей
	assert.Equal(t, float64(1), typeResponse["id"]) // JSON числа десериализуются как float64
	assert.Equal(t, "Йога", typeResponse["name"])
	assert.NotEmpty(t, typeResponse["created_at"])
}

// Тест на ошибку валидации при добавлении типа
func TestHandler_AddType_ValidationError(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем тестовый запрос с пустым именем
	body := []byte(`{"name": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/type", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addType(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Проверяем ответ
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "required")
}

// Тест на некорректный формат JSON при добавлении типа
func TestHandler_AddType_InvalidJSON(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем тестовый запрос с невалидным JSON
	body := []byte(`{"name": "Йога"`)
	req := httptest.NewRequest(http.MethodPost, "/admin/type", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addType(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на внутреннюю ошибку сервера при добавлении типа
func TestHandler_AddType_InternalError(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД с ошибкой
	sqlMock.ExpectQuery(`INSERT INTO content_types \(name, deleted\) VALUES \(\$1, FALSE\) RETURNING id, name, created_at`).
		WithArgs("Йога").
		WillReturnError(errors.New("database error"))

	// Создаем тестовый запрос
	body := []byte(`{"name": "Йога"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/type", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addType(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на успешное получение типа по ID
func TestHandler_GetType_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`SELECT id, name, created_at FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at"}).
			AddRow(1, "Йога", time.Now()))

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Get("/admin/type/{id}", h.getType)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/type/1", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Проверяем ответ
	var response dto.TypeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, "Йога", response.Name)
}

// Тест на получение несуществующего типа
func TestHandler_GetType_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`SELECT id, name, created_at FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Get("/admin/type/{id}", h.getType)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/type/999", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на невалидный ID типа
func TestHandler_GetType_InvalidID(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Get("/admin/type/{id}", h.getType)

	// Создаем тестовый запрос с невалидным ID
	req := httptest.NewRequest(http.MethodGet, "/admin/type/invalid", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на успешное получение всех типов
func TestHandler_GetTypes_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`SELECT id, name, created_at FROM content_types WHERE deleted IS NOT TRUE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at"}).
			AddRow(1, "Йога", time.Now()).
			AddRow(2, "Пилатес", time.Now()))

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/type", nil)
	w := httptest.NewRecorder()

	// Вызываем метод
	h.getTypes(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Проверяем ответ
	var response []dto.TypeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Йога", response[0].Name)
	assert.Equal(t, "Пилатес", response[1].Name)
}

// Тест на ошибку при получении списка типов
func TestHandler_GetTypes_Error(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД с ошибкой
	sqlMock.ExpectQuery(`SELECT id, name, created_at FROM content_types WHERE deleted IS NOT TRUE`).
		WillReturnError(errors.New("database error"))

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/type", nil)
	w := httptest.NewRecorder()

	// Вызываем метод
	h.getTypes(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на успешное обновление типа
func TestHandler_UpdateType_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectExec(`UPDATE content_types SET name = \$1 WHERE id = \$2 AND deleted IS NOT TRUE`).
		WithArgs("Обновленная йога", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Put("/admin/type/{id}", h.updateType)

	// Создаем тестовый запрос
	body := []byte(`{"name": "Обновленная йога"}`)
	req := httptest.NewRequest(http.MethodPut, "/admin/type/1", bytes.NewBuffer(body))
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
	assert.Contains(t, response["message"], "updated successfully")
}

// Тест на обновление несуществующего типа
func TestHandler_UpdateType_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectExec(`UPDATE content_types SET name = \$1 WHERE id = \$2 AND deleted IS NOT TRUE`).
		WithArgs("Обновленная йога", 999).
		WillReturnError(sql.ErrNoRows)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Put("/admin/type/{id}", h.updateType)

	// Создаем тестовый запрос
	body := []byte(`{"name": "Обновленная йога"}`)
	req := httptest.NewRequest(http.MethodPut, "/admin/type/999", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на ошибку валидации при обновлении типа
func TestHandler_UpdateType_ValidationError(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Put("/admin/type/{id}", h.updateType)

	// Создаем тестовый запрос с невалидным телом
	body := []byte(`{"name": ""}`)
	req := httptest.NewRequest(http.MethodPut, "/admin/type/1", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты - должно быть 400, так как имя пустое
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на успешное удаление типа
func TestHandler_DeleteType_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидания для новой логики с транзакцией
	sqlMock.ExpectBegin()

	// Ожидаем удаление связей с категориями
	sqlMock.ExpectExec(`DELETE FROM category_content_types WHERE content_type_id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2)) // Удаляем 2 связи

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectExec(`UPDATE content_types SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Ожидаем завершение транзакции
	sqlMock.ExpectCommit()

	// Настраиваем ожидания для redis (вызывается в горутине, но мы все равно можем проверить)
	redisMock.ExpectDel("categories:1").SetVal(1)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Delete("/admin/type/{id}", h.deleteType)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodDelete, "/admin/type/1", nil)
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

// Тест на удаление несуществующего типа
func TestHandler_DeleteType_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидания для новой логики с транзакцией
	sqlMock.ExpectBegin()

	// Ожидаем удаление связей с категориями
	sqlMock.ExpectExec(`DELETE FROM category_content_types WHERE content_type_id = \$1`).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0)) // Нет связей для удаления

	// Настраиваем ожидание запроса к БД для удаления типа, который не существует
	sqlMock.ExpectExec(`UPDATE content_types SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 затронутых строк

	// Ожидаем откат транзакции из-за ошибки ErrNoRows
	sqlMock.ExpectRollback()

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Delete("/admin/type/{id}", h.deleteType)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodDelete, "/admin/type/999", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на невалидный ID при удалении типа
func TestHandler_DeleteType_InvalidID(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Delete("/admin/type/{id}", h.deleteType)

	// Создаем тестовый запрос с невалидным ID
	req := httptest.NewRequest(http.MethodDelete, "/admin/type/invalid", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
