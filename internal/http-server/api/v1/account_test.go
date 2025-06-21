package v1

import (
	"database/sql"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/logdiscart"
	"github.com/langowen/bodybalance-backend/internal/storage/postgres"
	pgapi "github.com/langowen/bodybalance-backend/internal/storage/postgres/api"
	"github.com/langowen/bodybalance-backend/internal/storage/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestHandlerWithMocks создает тестовый обработчик с помощью sqlmock и redismock
func newTestHandlerWithMocks(t *testing.T) (*Handler, sqlmock.Sqlmock, redismock.ClientMock) {
	// Создаем мок для SQL
	db, sqlMock, err := sqlmock.New()
	require.NoError(t, err)

	// Создаем мок для Redis
	redisCli, redisMock := redismock.NewClientMock()

	// Создаем конфиг
	cfg := &config.Config{
		Redis: config.Redis{
			CacheTTL: time.Hour,
			Enable:   true,
		},
		Env: "test",
	}

	// Создаем redis storage с моком
	redisStorage := redis.NewStorage(redisCli, cfg)

	// Создаем postgres storage с моком
	pgPool := postgres.NewMockPgxPool(db)
	pgStorage := pgapi.New(pgPool.DB(), cfg)

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

func TestCheckAccount_FromCache(t *testing.T) {
	// Создаем моки и хендлер
	h, _, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для получения аккаунта из кэша
	acc := &response.AccountResponse{TypeID: 1, TypeName: "user"}
	accountData, err := json.Marshal(acc)
	require.NoError(t, err)

	redisMock.ExpectGet("account:user1").SetVal(string(accountData))

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/login?username=user1", nil)
	w := httptest.NewRecorder()

	// Вызываем метод
	h.checkAccount(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, redisMock.ExpectationsWereMet())

	var got response.AccountResponse
	err = json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.Equal(t, acc.TypeID, got.TypeID)
	assert.Equal(t, acc.TypeName, got.TypeName)
}

func TestCheckAccount_FromStorageAndCacheSet(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для отсутствия аккаунта в кэше
	redisMock.ExpectGet("account:user2").RedisNil()

	// Настраиваем мок SQL для получения аккаунта из БД
	// Исправлен SQL-запрос в соответствии с реальной реализацией
	sqlMock.ExpectQuery(`SELECT a.content_type_id, ct.name FROM accounts a JOIN content_types ct ON a.content_type_id = ct.id WHERE a.username = \$1 AND a.deleted IS NOT TRUE`).
		WithArgs("user2").
		WillReturnRows(sqlmock.NewRows([]string{"content_type_id", "name"}).
			AddRow(2, "premium"))

	// Важно: НЕ настраиваем ожидание вызова SetAccount, так как в реализации это делается асинхронно через горутину

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/login?username=user2", nil)
	w := httptest.NewRecorder()

	// Вызываем метод
	h.checkAccount(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	// Не проверяем ожидания Redis, так как вызов Set происходит асинхронно

	var got response.AccountResponse
	err := json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.Equal(t, int64(2), got.TypeID)
	assert.Equal(t, "premium", got.TypeName)
}

func TestCheckAccount_EmptyUsername(t *testing.T) {
	// Создаем моки и хендлер
	h, _, _ := newTestHandlerWithMocks(t)

	// Создаем тестовый запрос с пустым username
	req := httptest.NewRequest(http.MethodGet, "/login?username=", nil)
	w := httptest.NewRecorder()

	// Вызываем метод
	h.checkAccount(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCheckAccount_NotFound(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для отсутствия аккаунта в кэше
	redisMock.ExpectGet("account:nouser").RedisNil()

	// Настраиваем мок SQL для отсутствия аккаунта в БД
	sqlMock.ExpectQuery(`SELECT a.content_type_id, ct.name FROM accounts a JOIN content_types ct ON a.content_type_id = ct.id WHERE a.username = \$1 AND a.deleted IS NOT TRUE`).
		WithArgs("nouser").
		WillReturnError(sql.ErrNoRows)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/login?username=nouser", nil)
	w := httptest.NewRecorder()

	// Вызываем метод
	h.checkAccount(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestCheckAccount_StorageError(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для отсутствия аккаунта в кэше
	redisMock.ExpectGet("account:erruser").RedisNil()

	// Настраиваем мок SQL для ошибки БД
	sqlMock.ExpectQuery(`SELECT a.content_type_id, ct.name FROM accounts a JOIN content_types ct ON a.content_type_id = ct.id WHERE a.username = \$1 AND a.deleted IS NOT TRUE`).
		WithArgs("erruser").
		WillReturnError(sql.ErrConnDone)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/login?username=erruser", nil)
	w := httptest.NewRecorder()

	// Вызываем метод
	h.checkAccount(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestCheckAccount_RedisGetError(t *testing.T) {
	// Создаем моки и хендлер
	h, sqlMock, redisMock := newTestHandlerWithMocks(t)

	// Настраиваем мок Redis для ошибки при получении из кэша
	redisMock.ExpectGet("account:user3").SetErr(sql.ErrConnDone)

	// Настраиваем мок SQL для успешного получения из БД
	sqlMock.ExpectQuery(`SELECT a.content_type_id, ct.name FROM accounts a JOIN content_types ct ON a.content_type_id = ct.id WHERE a.username = \$1 AND a.deleted IS NOT TRUE`).
		WithArgs("user3").
		WillReturnRows(sqlmock.NewRows([]string{"content_type_id", "name"}).
			AddRow(3, "admin"))

	// Важно: НЕ настраиваем ожидание вызова SetAccount, так как в реализации это делается асинхронно через горутину

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/login?username=user3", nil)
	w := httptest.NewRecorder()

	// Вызываем метод
	h.checkAccount(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	// Не проверяем ожидания Redis, так как вызов Set происходит асинхронно

	var got response.AccountResponse
	err := json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.Equal(t, int64(3), got.TypeID)
	assert.Equal(t, "admin", got.TypeName)
}
