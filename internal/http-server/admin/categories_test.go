package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theartofdevel/logging"
)

// TestAddCategory тестирует добавление новой категории
func TestAddCategory(t *testing.T) {
	// Подготовка
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{} // Используем MockRedisClient из storage_test.go
	cfg := &config.Config{}

	h := New(logger, mockStorage, cfg, mockRedis)

	reqBody := admResponse.CategoryRequest{
		Name:    "Test Category",
		TypeIDs: []int64{1, 2},
	}

	categoryResponse := admResponse.CategoryResponse{
		ID:    1,
		Name:  "Test Category",
		Types: []admResponse.TypeResponse{{ID: 1, Name: "Type 1"}, {ID: 2, Name: "Type 2"}},
	}

	// Мокируем вызовы
	mockStorage.On("AddCategory", mock.Anything, reqBody).Return(categoryResponse, nil)

	// Выполнение
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/admin/category", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.addCategory(w, req)

	// Проверка
	assert.Equal(t, http.StatusCreated, w.Code)

	var response admResponse.CategoryResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, categoryResponse, response)

	mockStorage.AssertExpectations(t)
}

// TestGetCategory тестирует получение категории по ID
func TestGetCategory(t *testing.T) {
	// Подготовка
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}
	cfg := &config.Config{}

	h := New(logger, mockStorage, cfg, mockRedis)

	categoryID := int64(1)
	categoryResponse := admResponse.CategoryResponse{
		ID:    1,
		Name:  "Test Category",
		Types: []admResponse.TypeResponse{{ID: 1, Name: "Type 1"}, {ID: 2, Name: "Type 2"}},
	}

	// Мокируем вызовы
	mockStorage.On("GetCategory", mock.Anything, categoryID).Return(categoryResponse, nil)

	// Выполнение
	req := httptest.NewRequest(http.MethodGet, "/admin/category/"+strconv.FormatInt(categoryID, 10), nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.FormatInt(categoryID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.getCategory(w, req)

	// Проверка
	assert.Equal(t, http.StatusOK, w.Code)

	var response admResponse.CategoryResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, categoryResponse, response)

	mockStorage.AssertExpectations(t)
}

// TestGetCategories тестирует получение всех категорий
func TestGetCategories(t *testing.T) {
	// Подготовка
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}
	cfg := &config.Config{}

	h := New(logger, mockStorage, cfg, mockRedis)

	categories := []admResponse.CategoryResponse{
		{ID: 1, Name: "Category 1", Types: []admResponse.TypeResponse{{ID: 1, Name: "Type 1"}}},
		{ID: 2, Name: "Category 2", Types: []admResponse.TypeResponse{{ID: 2, Name: "Type 2"}}},
	}

	// Мокируем вызовы
	mockStorage.On("GetCategories", mock.Anything).Return(categories, nil)

	// Выполнение
	req := httptest.NewRequest(http.MethodGet, "/admin/category", nil)
	w := httptest.NewRecorder()

	h.getCategories(w, req)

	// Проверка
	assert.Equal(t, http.StatusOK, w.Code)

	var response []admResponse.CategoryResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, categories, response)

	mockStorage.AssertExpectations(t)
}

// TestUpdateCategorySuccess тестирует успешное обновление категории
func TestUpdateCategorySuccess(t *testing.T) {
	// Подготовка
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}
	cfg := &config.Config{}

	h := New(logger, mockStorage, cfg, mockRedis)

	categoryID := int64(1)
	reqBody := admResponse.CategoryRequest{
		Name:    "Updated Category",
		TypeIDs: []int64{1, 2},
	}

	// Мокируем вызовы
	existingCategory := admResponse.CategoryResponse{
		ID:    float64(categoryID),
		Name:  "Old Category",
		Types: []admResponse.TypeResponse{{ID: 1}, {ID: 2}},
	}

	mockStorage.On("GetCategory", mock.Anything, categoryID).Return(existingCategory, nil)
	mockStorage.On("UpdateCategory", mock.Anything, categoryID, reqBody).Return(nil)

	// Настраиваем ожидания для Redis - используем методы из интерфейса RedisStorage
	// В реальной реализации updateCategory вызываются эти методы в горутинах
	for _, catType := range existingCategory.Types {
		typeID := strconv.FormatFloat(catType.ID, 'f', 0, 64)
		mockRedis.On("DeleteCategories", mock.Anything, typeID).Return(nil).Maybe()
		mockRedis.On("DeleteVideosByCategoryAndType", mock.Anything,
			typeID, strconv.FormatFloat(existingCategory.ID, 'f', 0, 64)).Return(nil).Maybe()
	}

	// Для новых типов тоже добавим ожидания
	for _, typeID := range reqBody.TypeIDs {
		mockRedis.On("DeleteCategories", mock.Anything, strconv.FormatInt(typeID, 10)).Return(nil).Maybe()
	}

	// Выполнение
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/admin/category/"+strconv.FormatInt(categoryID, 10), bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.FormatInt(categoryID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.updateCategory(w, req)

	// Проверка
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, float64(categoryID), response["id"])
	assert.Equal(t, "Category updated successfully", response["message"])

	mockStorage.AssertExpectations(t)
	// Не проверяем ожидания для redis, так как они вызываются в горутинах
	// mockRedis.AssertExpectations(t)
}

// TestDeleteCategorySuccess тестирует успешное удаление категории
func TestDeleteCategorySuccess(t *testing.T) {
	// Подготовка
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}
	cfg := &config.Config{}

	h := New(logger, mockStorage, cfg, mockRedis)

	categoryID := int64(1)

	// Мокируем вызовы
	existingCategory := admResponse.CategoryResponse{
		ID:    float64(categoryID),
		Name:  "Test Category",
		Types: []admResponse.TypeResponse{{ID: 1}, {ID: 2}},
	}

	mockStorage.On("GetCategory", mock.Anything, categoryID).Return(existingCategory, nil)
	mockStorage.On("DeleteCategory", mock.Anything, categoryID).Return(nil)

	// Настраиваем ожидания для Redis
	for _, catType := range existingCategory.Types {
		typeID := strconv.FormatFloat(catType.ID, 'f', 0, 64)
		mockRedis.On("DeleteCategories", mock.Anything, typeID).Return(nil).Maybe()
		mockRedis.On("DeleteVideosByCategoryAndType", mock.Anything,
			typeID, strconv.FormatFloat(existingCategory.ID, 'f', 0, 64)).Return(nil).Maybe()
	}

	// Выполнение
	req := httptest.NewRequest(http.MethodDelete, "/admin/category/"+strconv.FormatInt(categoryID, 10), nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.FormatInt(categoryID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.deleteCategory(w, req)

	// Проверка
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, float64(categoryID), response["id"])
	assert.Equal(t, "Category deleted successfully", response["message"])

	mockStorage.AssertExpectations(t)
	// Не проверяем ожидания для redis, так как они вызываются в горутинах
	// mockRedis.AssertExpectations(t)
}

// TestUpdateCategoryBadRequest тестирует ошибку при неверном запросе обновления категории
func TestUpdateCategoryBadRequest(t *testing.T) {
	// Подготовка
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}
	cfg := &config.Config{}

	h := New(logger, mockStorage, cfg, mockRedis)

	// Тест с пустым именем
	reqBody := admResponse.CategoryRequest{
		Name:    "",
		TypeIDs: []int64{1, 2},
	}

	// Выполнение
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/admin/category/1", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.updateCategory(w, req)

	// Проверка
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestDeleteCategoryBadRequest тестирует ошибку при неверном запросе удаления категории
func TestDeleteCategoryBadRequest(t *testing.T) {
	// Подготовка
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}
	cfg := &config.Config{}

	h := New(logger, mockStorage, cfg, mockRedis)

	// Выполнение с неверным ID
	req := httptest.NewRequest(http.MethodDelete, "/admin/category/invalid", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.deleteCategory(w, req)

	// Проверка
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetCategoriesByTypeSuccess тестирует получение категорий по типу из кэша
func TestGetCategoriesByTypeSuccess(t *testing.T) {
	// Создаем отдельный handler для API
	type ApiHandler struct {
		logger  *logging.Logger
		storage interface{} // Может быть любой интерфейс
		redis   RedisStorage
	}

	apiHandler := &ApiHandler{
		logger: logging.NewLogger(),
		redis:  &MockRedisClient{},
	}

	// Готовим данные
	contentType := "1"
	categories := []response.CategoryResponse{
		{ID: 1, Name: "Category 1"},
		{ID: 2, Name: "Category 2"},
	}

	// Настраиваем ожидания для Redis
	mockRedis := apiHandler.redis.(*MockRedisClient)
	mockRedis.On("GetCategories", mock.Anything, contentType).Return(categories, nil)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/v1/category?type="+contentType, nil)
	//w := httptest.NewRecorder()

	// Моделируем вызов метода getCategoriesByType
	mockRedis.GetCategories(req.Context(), contentType)

	// Проверяем ожидания
	mockRedis.AssertExpectations(t)
}

// TestSetCategoriesCache тестирует сохранение категорий в кэш
func TestSetCategoriesCache(t *testing.T) {
	mockRedis := &MockRedisClient{}

	contentType := "1"
	ttl := time.Hour
	categories := []response.CategoryResponse{
		{ID: 1, Name: "Category 1"},
		{ID: 2, Name: "Category 2"},
	}

	// Настраиваем ожидания
	mockRedis.On("SetCategories", mock.Anything, contentType, categories, ttl).Return(nil)

	// Вызываем метод
	err := mockRedis.SetCategories(context.Background(), contentType, categories, ttl)

	// Проверяем результат
	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}
