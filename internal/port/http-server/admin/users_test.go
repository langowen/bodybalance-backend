package admin

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin/admResponse"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тест на успешное добавление нового пользователя
func TestHandler_AddUser_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД для добавления пользователя
	sqlMock.ExpectQuery(`INSERT INTO accounts \(username, content_type_id, admin, password, deleted\) VALUES \(\$1, \$2, \$3, \$4, FALSE\) RETURNING id, username, content_type_id, \(SELECT name FROM content_types WHERE id = \$2\), admin, created_at`).
		WithArgs("testuser", int64(1), false, "password123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "content_type_id", "content_type_name", "admin", "created_at"}).
			AddRow(1, "testuser", "1", "Йога", false, time.Now()))

	// Создаем тестовый запрос
	req := admResponse.UserRequest{
		Username:      "testuser",
		Password:      "password123",
		ContentTypeID: 1,
		Admin:         false,
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addUser(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Проверяем ответ
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "successfully")
	assert.Equal(t, "testuser", response["username"])
}

// Тест на добавление пользователя с существующим именем
func TestHandler_AddUser_Duplicate(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД с ошибкой дубликата
	sqlMock.ExpectQuery(`INSERT INTO accounts \(username, content_type_id, admin, password, deleted\) VALUES \(\$1, \$2, \$3, \$4, FALSE\) RETURNING id, username, content_type_id, \(SELECT name FROM content_types WHERE id = \$2\), admin, created_at`).
		WithArgs("existinguser", int64(1), false, "password123").
		WillReturnError(errors.New("duplicate key value violates unique constraint"))

	// Создаем тестовый запрос
	req := admResponse.UserRequest{
		Username:      "existinguser",
		Password:      "password123",
		ContentTypeID: 1,
		Admin:         false,
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addUser(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на ошибку валидации при создании пользователя
func TestHandler_AddUser_ValidationError(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем тестовый запрос с пустыми обязательными полями
	req := admResponse.UserRequest{
		Username:      "", // Пустое имя пользователя
		Password:      "", // Пустой пароль
		ContentTypeID: 1,
		Admin:         false,
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addUser(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на невалидный формат JSON
func TestHandler_AddUser_InvalidJSON(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем тестовый запрос с невалидным JSON
	httpReq := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewBuffer([]byte(`{"username": "testuser", "password": `)))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addUser(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на успешное получение пользователя по ID
func TestHandler_GetUser_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`SELECT a\.id, a\.username, a\.content_type_id, ct\.name, a\.admin, a\.created_at FROM accounts a LEFT JOIN content_types ct ON a\.content_type_id = ct\.id WHERE a\.id = \$1 AND a\.deleted IS NOT TRUE`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "content_type_id", "content_type_name", "admin", "created_at"}).
			AddRow(1, "testuser", "1", "Йога", false, time.Now()))

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Get("/admin/users/{id}", h.getUser)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/users/1", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Проверяем ответ
	var response admResponse.UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, "testuser", response.Username)
}

// Тест на получение несуществующего пользователя
func TestHandler_GetUser_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`SELECT a\.id, a\.username, a\.content_type_id, ct\.name, a\.admin, a\.created_at FROM accounts a LEFT JOIN content_types ct ON a\.content_type_id = ct\.id WHERE a\.id = \$1 AND a\.deleted IS NOT TRUE`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Get("/admin/users/{id}", h.getUser)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/users/999", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на невалидный ID пользователя
func TestHandler_GetUser_InvalidID(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Get("/admin/users/{id}", h.getUser)

	// Создаем тестовый запрос с невалидным ID
	req := httptest.NewRequest(http.MethodGet, "/admin/users/invalid", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на успешное получение всех пользователей
func TestHandler_GetUsers_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`SELECT a\.id, a\.username, a\.content_type_id, ct\.name, a\.admin, a\.created_at FROM accounts a LEFT JOIN content_types ct ON a\.content_type_id = ct\.id WHERE a\.deleted IS NOT TRUE ORDER BY a\.id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "content_type_id", "content_type_name", "admin", "created_at"}).
			AddRow(1, "user1", "1", "Йога", false, time.Now()).
			AddRow(2, "user2", "2", "Пилатес", false, time.Now()))

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	w := httptest.NewRecorder()

	// Вызываем метод
	h.getUsers(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Проверяем ответ
	var response []admResponse.UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "user1", response[0].Username)
	assert.Equal(t, "user2", response[1].Username)
}

// Тест на ошибку при получении списка пользователей
func TestHandler_GetUsers_Error(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД с ошибкой
	sqlMock.ExpectQuery(`SELECT a\.id, a\.username, a\.content_type_id, ct\.name, a\.admin, a\.created_at FROM accounts a LEFT JOIN content_types ct ON a\.content_type_id = ct\.id WHERE a\.deleted IS NOT TRUE ORDER BY a\.id`).
		WillReturnError(errors.New("database error"))

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	w := httptest.NewRecorder()

	// Вызываем метод
	h.getUsers(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на успешное обновление пользователя
func TestHandler_UpdateUser_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectExec(`UPDATE accounts SET username = \$1, content_type_id = \$2, admin = \$3, password = \$4 WHERE id = \$5 AND deleted IS NOT TRUE`).
		WithArgs("updateduser", int64(2), true, "newpassword", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Настраиваем ожидания для redis (вызывается в горутине)
	redisMock.ExpectDel("account:updateduser").SetVal(1)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Put("/admin/users/{id}", h.updateUser)

	// Создаем тестовый запрос
	req := admResponse.UserRequest{
		Username:      "updateduser",
		Password:      "newpassword",
		ContentTypeID: 2,
		Admin:         true,
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPut, "/admin/users/1", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, httpReq)

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

// Тест на обновление несуществующего пользователя
func TestHandler_UpdateUser_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectExec(`UPDATE accounts SET username = \$1, content_type_id = \$2, admin = \$3, password = \$4 WHERE id = \$5 AND deleted IS NOT TRUE`).
		WithArgs("updateduser", int64(2), true, "newpassword", int64(999)).
		WillReturnError(sql.ErrNoRows)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Put("/admin/users/{id}", h.updateUser)

	// Создаем тестовый запрос
	req := admResponse.UserRequest{
		Username:      "updateduser",
		Password:      "newpassword",
		ContentTypeID: 2,
		Admin:         true,
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPut, "/admin/users/999", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на невалидный формат JSON при обновлении
func TestHandler_UpdateUser_InvalidJSON(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Put("/admin/users/{id}", h.updateUser)

	// Создаем тестовый запрос с невалидным JSON
	httpReq := httptest.NewRequest(http.MethodPut, "/admin/users/1", bytes.NewBuffer([]byte(`{"username": "testuser", "password": `)))
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на успешное удаление пользователя
func TestHandler_DeleteUser_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД для удаления пользователя
	sqlMock.ExpectExec(`UPDATE accounts SET deleted = TRUE WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Настраиваем ожидание запроса к БД для получения имени пользователя (в горутине)
	sqlMock.ExpectQuery(`SELECT a.id, a.username, a.content_type_id, ct.name FROM accounts`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "content_type_id", "content_type_name", "admin", "created_at"}).
			AddRow(1, "testuser", "1", "Йога", false, time.Now()))

	// Настраиваем ожидания для redis (вызывается в горутине)
	redisMock.ExpectDel("account:testuser").SetVal(1)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Delete("/admin/users/{id}", h.deleteUser)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodDelete, "/admin/users/1", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем ответ
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(1), response["id"])
	assert.Contains(t, response["message"], "deleted successfully")
}

// Тест на удаление несуществующего пользователя
func TestHandler_DeleteUser_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectExec(`UPDATE accounts SET deleted = TRUE WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Delete("/admin/users/{id}", h.deleteUser)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodDelete, "/admin/users/999", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на невалидный ID пользователя при удалении
func TestHandler_DeleteUser_InvalidID(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Delete("/admin/users/{id}", h.deleteUser)

	// Создаем тестовый запрос с невалидным ID
	req := httptest.NewRequest(http.MethodDelete, "/admin/users/invalid", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
