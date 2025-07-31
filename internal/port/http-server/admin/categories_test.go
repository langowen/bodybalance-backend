package admin

import (
	"bytes"
	"context"
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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// RedisMock - мок для redis хранилища
type RedisMock struct {
	mock.Mock
}

// InvalidateVideosCache - мок метода инвалидации кэша видео
func (m *RedisMock) InvalidateVideosCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// InvalidateCategoriesCache - мок метода инвалидации кэша категорий
func (m *RedisMock) InvalidateCategoriesCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// InvalidateAccountsCache - мок метода инвалидации кэша аккаунтов
func (m *RedisMock) InvalidateAccountsCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Тест на успешное добавление категории
func TestHandler_AddCategory_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД для добавления категории
	sqlMock.ExpectBegin()

	// Имитируем вставку категории
	sqlMock.ExpectQuery(`INSERT INTO categories \(name, img_url, deleted\) VALUES \(\$1, \$2, FALSE\) RETURNING id, name, img_url, created_at`).
		WithArgs("Test Category", "image.jpg").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "img_url", "created_at"}).
			AddRow(1, "Test Category", "image.jpg", time.Now()))

	// Настраиваем ожидание для добавления связей с типами
	sqlMock.ExpectExec(`INSERT INTO category_content_types \(category_id, content_type_id\) VALUES \(\$1, \$2\) ON CONFLICT DO NOTHING`).
		WithArgs(1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Настраиваем ожидание для запроса связанных типов
	sqlMock.ExpectQuery(`SELECT ct\.id, ct\.name FROM content_types ct JOIN category_content_types cct ON ct\.id = cct\.content_type_id WHERE cct\.category_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Yoga"))

	sqlMock.ExpectCommit()

	// Создаем тестовый запрос на добавление категории
	req := admResponse.CategoryRequest{
		Name:    "Test Category",
		ImgURL:  "image.jpg", // Используем только имя файла без URL
		TypeIDs: []int64{1},
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/admin/category", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addCategory(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Проверяем ответ
	var response admResponse.CategoryResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, "Test Category", response.Name)
	assert.Equal(t, "image.jpg", response.ImgURL)
}

// Тест на получение категории по ID
func TestHandler_GetCategory_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`SELECT id, name, img_url, created_at FROM categories WHERE id = \$1 AND deleted IS NOT TRUE`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "img_url", "created_at"}).
			AddRow(1, "Test Category", "https://example.com/image.jpg", time.Now()))

	// Настраиваем запрос для получения типов категории
	sqlMock.ExpectQuery(`SELECT ct.id, ct.name FROM content_types ct JOIN category_content_types`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Yoga"))

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Get("/admin/category/{id}", h.getCategory)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/category/1", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Проверяем ответ
	var response admResponse.CategoryResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, "Test Category", response.Name)
}

// Тест на получение всех категорий
func TestHandler_GetCategories_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`SELECT id, name, img_url, created_at FROM categories WHERE deleted IS NOT TRUE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "img_url", "created_at"}).
			AddRow(1, "Category 1", "https://example.com/image1.jpg", time.Now()).
			AddRow(2, "Category 2", "https://example.com/image2.jpg", time.Now()))

	// Настраиваем запросы для получения типов категорий
	sqlMock.ExpectQuery(`SELECT ct\.id, ct\.name FROM content_types ct JOIN category_content_types cct ON ct\.id = cct\.content_type_id WHERE cct\.category_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Yoga"))

	sqlMock.ExpectQuery(`SELECT ct\.id, ct\.name FROM content_types ct JOIN category_content_types cct ON ct\.id = cct\.content_type_id WHERE cct\.category_id = \$1`).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(2, "Pilates"))

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/category", nil)
	w := httptest.NewRecorder()

	// Вызываем метод
	h.getCategories(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Проверяем ответ
	var response []admResponse.CategoryResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Category 1", response[0].Name)
	assert.Equal(t, "Category 2", response[1].Name)
}

// Тест на обновление категории
func TestHandler_UpdateCategory_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД для обновления категории
	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(`UPDATE categories SET name = \$1, img_url = \$2 WHERE id = \$3 AND deleted IS NOT TRUE`).
		WithArgs("Updated Category", "updated.jpg", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Настраиваем ожидание для удаления существующих связей
	sqlMock.ExpectExec(`DELETE FROM category_content_types WHERE category_id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Настраиваем ожидание для добавления новых связей
	sqlMock.ExpectExec(`INSERT INTO category_content_types`).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	sqlMock.ExpectCommit()

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Put("/admin/category/{id}", h.updateCategory)

	// Создаем тестовый запрос
	req := admResponse.CategoryRequest{
		Name:    "Updated Category",
		ImgURL:  "updated.jpg", // Используем только имя файла без URL
		TypeIDs: []int64{2},
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPut, "/admin/category/1", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, httpReq)

	// Проверяем результаты - не ждем запроса в removeCategoryCache, так как он выполняется в горутине
	assert.Equal(t, http.StatusOK, w.Code)
}

// Тест на удаление категории
func TestHandler_DeleteCategory_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидания для новой логики с транзакцией
	sqlMock.ExpectBegin()

	// Ожидаем удаление связей с типами контента
	sqlMock.ExpectExec(`DELETE FROM category_content_types WHERE category_id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Ожидаем удаление связей с видео
	sqlMock.ExpectExec(`DELETE FROM video_categories WHERE category_id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// Настраиваем ожидание запроса к БД для удаления категории
	sqlMock.ExpectExec(`UPDATE categories SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Ожидаем завершение транзакции
	sqlMock.ExpectCommit()

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Delete("/admin/category/{id}", h.deleteCategory)

	// Создаем тестовый запрос
	httpReq := httptest.NewRequest(http.MethodDelete, "/admin/category/1", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, httpReq)

	// Проверяем результаты - не ждем запроса в removeCategoryCache, так как он выполняется в горутине
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на ошибку валидации при добавлении категории
func TestHandler_AddCategory_ValidationError(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем тестовый запрос с пустыми полями
	req := admResponse.CategoryRequest{
		Name:    "",
		TypeIDs: []int64{},
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/admin/category", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addCategory(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на ошибку валидации JSON при добавлении категории
func TestHandler_AddCategory_InvalidJSON(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем тестовый запрос с некорректным JSON
	body := []byte(`{"name": "Test", "type_ids": [1,}`)
	httpReq := httptest.NewRequest(http.MethodPost, "/admin/category", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.addCategory(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на случай, когда категория не найдена
func TestHandler_GetCategory_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`SELECT id, name, img_url, created_at FROM categories WHERE id = \$1 AND deleted IS NOT TRUE`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Get("/admin/category/{id}", h.getCategory)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/category/999", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на внутреннюю ошибку сервера при получении категорий
func TestHandler_GetCategories_InternalError(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД с возвратом ошибки
	sqlMock.ExpectQuery(`SELECT id, name, img_url, created_at FROM categories WHERE deleted IS NOT TRUE`).
		WillReturnError(errors.New("database error"))

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/category", nil)
	w := httptest.NewRecorder()

	// Вызываем метод
	h.getCategories(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на ошибку при обновлении категории с невалидным ID
func TestHandler_UpdateCategory_InvalidID(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем роутер с невалидным параметром id
	r := chi.NewRouter()
	r.Put("/admin/category/{id}", h.updateCategory)

	// Создаем тестовый запрос
	req := admResponse.CategoryRequest{
		Name:    "Updated Category",
		TypeIDs: []int64{1},
	}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPut, "/admin/category/invalid", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на удаление несуществующей категории
func TestHandler_DeleteCategory_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидания для новой логики с транзакцией
	sqlMock.ExpectBegin()

	// Ожидаем удаление связей с типами контента
	sqlMock.ExpectExec(`DELETE FROM category_content_types WHERE category_id = \$1`).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Ожидаем удаление связей с видео
	sqlMock.ExpectExec(`DELETE FROM video_categories WHERE category_id = \$1`).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Настраиваем ожидание запроса к БД для удаления категории
	sqlMock.ExpectExec(`UPDATE categories SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Ожидаем откат транзакции из-за ошибки ErrNoRows
	sqlMock.ExpectRollback()

	// Создаем роутер с параметром id
	r := chi.NewRouter()
	r.Delete("/admin/category/{id}", h.deleteCategory)

	// Создаем тестовый запрос
	httpReq := httptest.NewRequest(http.MethodDelete, "/admin/category/999", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос через роутер
	r.ServeHTTP(w, httpReq)

	// Проверяем результаты
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestHandler_RemoveCache_Success проверяет правильность инвалидации кэша
func TestHandler_RemoveCache_Success(t *testing.T) {
	// Создаем моки и хендлер
	h, _, redisMockClient := newTestAuthHandlerWithMocks(t)

	// Настройка ожиданий redis для каждого шаблона
	// Для videos:*
	redisMockClient.ExpectScan(uint64(0), "videos:*", int64(100)).SetVal([]string{"videos:1:1", "videos:2:1"}, 0)
	redisMockClient.ExpectDel("videos:1:1", "videos:2:1").SetVal(2)

	// Для categories:*
	redisMockClient.ExpectScan(uint64(0), "categories:*", int64(100)).SetVal([]string{"categories:1", "categories:2"}, 0)
	redisMockClient.ExpectDel("categories:1", "categories:2").SetVal(2)

	// Для account:*
	redisMockClient.ExpectScan(uint64(0), "account:*", int64(100)).SetVal([]string{"account:user1", "account:user2"}, 0)
	redisMockClient.ExpectDel("account:user1", "account:user2").SetVal(2)

	// Вызываем метод
	h.removeCache("test.op")

	// Даем время горутине завершиться
	time.Sleep(50 * time.Millisecond)

	// Проверяем, что все ожидания были выполнены
	if err := redisMockClient.ExpectationsWereMet(); err != nil {
		t.Errorf("есть невыполненные ожидания: %s", err)
	}
}
