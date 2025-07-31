package admin

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/internal/adapter/storage/postgres"
	pgadmin "github.com/langowen/bodybalance-backend/internal/adapter/storage/postgres/admin"
	"github.com/langowen/bodybalance-backend/internal/adapter/storage/redis"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/logdiscart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestAuthHandlerWithMocks создает тестовый обработчик с помощью sqlmock и redismock
func newTestAuthHandlerWithMocks(t *testing.T) (*Handler, sqlmock.Sqlmock, redismock.ClientMock) {
	// Создаем мок для SQL
	db, sqlMock, err := sqlmock.New()
	require.NoError(t, err)

	// Создаем мок для redis
	redisCli, redisMock := redismock.NewClientMock()

	// Создаем postgres storage с моком
	pgPool := postgres.NewMockPgxPool(db)
	pgStorage := pgadmin.New(pgPool.DB())

	// Создаем конфиг
	cfg := &config.Config{
		HTTPServer: config.HTTPServer{
			TokenTTL:   time.Hour,
			SigningKey: "test-signing-key",
		},
		Redis: config.Redis{
			Enable: true,
		},
		Env: "test",
	}

	// Создаем redis storage с моком
	redisStorage := redis.NewStorage(redisCli, cfg)

	// Создаем логгер
	logger := logdiscart.NewDiscardLogger()

	// Создаем хендлер
	handler := &Handler{
		storage: pgStorage,
		redis:   redisStorage,
		logger:  logger,
		cfg:     cfg,
	}

	return handler, sqlMock, redisMock
}

// Тест на успешную авторизацию администратора
func TestHandler_Signing_Success(t *testing.T) {
	// Создаем моки и хендлер с помощью вспомогательной функции
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`SELECT username, password, admin FROM accounts WHERE username = \$1 AND password = \$2 AND admin = TRUE AND deleted IS NOT TRUE`).
		WithArgs("admin", "password").
		WillReturnRows(sqlmock.NewRows([]string{"username", "password", "admin"}).
			AddRow("admin", "password", true))

	// Создаем тестовый запрос
	body := []byte(`{"login":"admin", "password":"password"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.signing(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())

	// Проверяем, что cookie установлен
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies)
	assert.Equal(t, "token", cookies[0].Name)
}

// Тест на ошибочные учетные данные
func TestHandler_Signing_InvalidCredentials(t *testing.T) {
	// Создаем моки и хендлер с помощью вспомогательной функции
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`SELECT username, password, admin FROM accounts WHERE username = \$1 AND password = \$2 AND admin = TRUE AND deleted IS NOT TRUE`).
		WithArgs("admin", "wrong").
		WillReturnError(sql.ErrNoRows)

	// Создаем тестовый запрос
	body := []byte(`{"login":"admin", "password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Вызываем метод
	h.signing(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на выход из системы
func TestHandler_Logout_Success(t *testing.T) {
	// Создаем моки и хендлер с помощью вспомогательной функции
	h, _, _ := newTestAuthHandlerWithMocks(t)

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
	// Создаем моки и хендлер с помощью вспомогательной функции
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`SELECT username, password, admin FROM accounts WHERE username = \$1 AND password = \$2 AND admin = TRUE AND deleted IS NOT TRUE`).
		WithArgs("admin", "password").
		WillReturnError(errors.New("db error"))

	body := []byte(`{"login":"admin", "password":"password"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.signing(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// Тест на невалидный JSON
func TestHandler_Signing_InvalidJSON(t *testing.T) {
	// Создаем моки и хендлер с помощью вспомогательной функции
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем тестовый запрос с невалидным JSON
	body := []byte(`{"login": "admin", "password": }`)
	req := httptest.NewRequest(http.MethodPost, "/admin/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.signing(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на пустые учетные данные
func TestHandler_Signing_EmptyCredentials(t *testing.T) {
	// Создаем моки и хендлер с помощью вспомогательной функции
	h, _, _ := newTestAuthHandlerWithMocks(t)

	// Создаем тестовый запрос с пустыми данными
	body := []byte(`{"login": "", "password": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.signing(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на пользователя без прав администратора
func TestHandler_Signing_NotAdmin(t *testing.T) {
	// Создаем моки и хендлер с помощью вспомогательной функции
	h, sqlMock, _ := newTestAuthHandlerWithMocks(t)

	// Настраиваем ожидание запроса к БД
	sqlMock.ExpectQuery(`SELECT username, password, admin FROM accounts WHERE username = \$1 AND password = \$2 AND admin = TRUE AND deleted IS NOT TRUE`).
		WithArgs("user", "password").
		WillReturnRows(sqlmock.NewRows([]string{"username", "password", "admin"}).
			AddRow("user", "password", false))

	// Создаем тестовый запрос
	body := []byte(`{"login":"user", "password":"password"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.signing(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
