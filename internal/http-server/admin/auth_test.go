package admin

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/storage/postgres/admin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theartofdevel/logging"
)

// Тест на успешную авторизацию администратора
func TestHandler_Signing_Success(t *testing.T) {
	// Создаем моки
	logger := logging.NewLogger()
	storage := &MockAdmStorage{}
	cfg := &config.Config{
		HTTPServer: config.HTTPServer{
			TokenTTL:   time.Hour,
			SigningKey: "test-signing-key",
		},
		Env: "test",
	}

	// Настраиваем ожидания
	storage.On("GetAdminUser", mock.Anything, "admin", "password").Return(&admin.AdmUser{
		Username: "admin",
		Password: "password",
		IsAdmin:  true,
	}, nil)

	// Создаем хендлер
	h := New(logger, storage, cfg, nil)

	// Создаем тестовый запрос
	body := []byte(`{"login":"admin", "password":"password"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.signing(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	storage.AssertExpectations(t)

	// Проверяем, что cookie установлен
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies)
	assert.Equal(t, "token", cookies[0].Name)
}

// Тест на ошибочные учетные данные
func TestHandler_Signing_InvalidCredentials(t *testing.T) {
	// Создаем моки
	logger := logging.NewLogger()
	storage := &MockAdmStorage{}
	cfg := &config.Config{
		HTTPServer: config.HTTPServer{
			TokenTTL:   time.Hour,
			SigningKey: "test-signing-key",
		},
		Env: "test",
	}

	// Настраиваем ожидания
	storage.On("GetAdminUser", mock.Anything, "admin", "wrong").Return(nil, sql.ErrNoRows)

	// Создаем хендлер
	h := New(logger, storage, cfg, nil)

	// Создаем тестовый запрос
	body := []byte(`{"login":"admin", "password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.signing(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	storage.AssertExpectations(t)
}

// Тест на выход из системы
func TestHandler_Logout_Success(t *testing.T) {
	// Создаем моки
	logger := logging.NewLogger()
	cfg := &config.Config{Env: "test"}

	h := New(logger, nil, cfg, nil)
	req := httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
	w := httptest.NewRecorder()

	h.logout(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	cookies := w.Result().Cookies()
	assert.Equal(t, "token", cookies[0].Name)
	assert.Equal(t, -1, cookies[0].MaxAge)
}

// Тест на внутреннюю ошибку сервера
func TestHandler_Signing_InternalError(t *testing.T) {
	// Создаем моки
	logger := logging.NewLogger()
	storage := &MockAdmStorage{}
	cfg := &config.Config{
		HTTPServer: config.HTTPServer{
			TokenTTL:   time.Hour,
			SigningKey: "test-signing-key",
		},
		Env: "test",
	}

	storage.On("GetAdminUser", mock.Anything, "admin", "password").Return(nil, errors.New("db error"))

	h := New(logger, storage, cfg, nil)
	body := []byte(`{"login":"admin", "password":"password"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.signing(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	storage.AssertExpectations(t)
}

// Тест на невалидный JSON
func TestHandler_Signing_InvalidJSON(t *testing.T) {
	// Создаем моки
	logger := logging.NewLogger()
	storage := &MockAdmStorage{}
	cfg := &config.Config{
		HTTPServer: config.HTTPServer{
			TokenTTL:   time.Hour,
			SigningKey: "test-signing-key",
		},
		Env: "test",
	}

	h := New(logger, storage, cfg, nil)

	// Создаем тестовый запрос с невалидным JSON
	body := []byte(`{"login": "admin", "password": }`)
	req := httptest.NewRequest(http.MethodPost, "/admin/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.signing(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на пустые учетные данные
func TestHandler_Signing_EmptyCredentials(t *testing.T) {
	// Создаем моки
	logger := logging.NewLogger()
	storage := &MockAdmStorage{}
	cfg := &config.Config{
		HTTPServer: config.HTTPServer{
			TokenTTL:   time.Hour,
			SigningKey: "test-signing-key",
		},
		Env: "test",
	}

	h := New(logger, storage, cfg, nil)

	// Создаем тестовый запрос с пустыми данными
	body := []byte(`{"login": "", "password": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.signing(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на пользователя без прав администратора
func TestHandler_Signing_NotAdmin(t *testing.T) {
	// Создаем моки
	logger := logging.NewLogger()
	storage := &MockAdmStorage{}
	cfg := &config.Config{
		HTTPServer: config.HTTPServer{
			TokenTTL:   time.Hour,
			SigningKey: "test-signing-key",
		},
		Env: "test",
	}

	// Настраиваем ожидания
	storage.On("GetAdminUser", mock.Anything, "user", "password").Return(&admin.AdmUser{
		Username: "user",
		Password: "password",
		IsAdmin:  false,
	}, nil)

	h := New(logger, storage, cfg, nil)

	// Создаем тестовый запрос
	body := []byte(`{"login":"user", "password":"password"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.signing(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	storage.AssertExpectations(t)
}
