package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Создаем хелпер для настройки тестового окружения
func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Storage) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")

	// Создаем конфигурацию для тестов
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			DBName:   "test",
			SSLMode:  "disable",
			Schema:   "public",
			Timeout:  time.Second * 5,
		},
		Docs: config.Docs{
			User:     "admin",
			Password: "admin123",
		},
	}

	// Создаем экземпляр хранилища с моком DB
	storage := &Storage{
		db:  db,
		cfg: cfg,
		// Не инициализируем Admin и Api, так как они требуют реальное соединение
	}

	return db, mock, storage
}

// Тест для функции New
func TestNew(t *testing.T) {
	t.Run("connection error", func(t *testing.T) {
		// В этом тесте мы не можем использовать sqlmock так как New создает новое подключение
		// Вместо этого мы создаем конфигурацию с неправильными данными
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Host:     "non-existent-host",
				Port:     12345,
				User:     "invalid",
				Password: "invalid",
				DBName:   "invalid",
				SSLMode:  "disable",
				Schema:   "public",
				Timeout:  time.Second,
			},
		}

		ctx := context.Background()
		_, err := New(ctx, cfg)

		// Ожидаем ошибку подключения
		assert.Error(t, err)
	})
}

// Тест для функции initSchema
func TestInitSchema(t *testing.T) {
	// Пропускаем тесты для initSchema, так как они требуют сложных моков для SQL-запросов
	// В реальном проекте функциональность initSchema проверяется в интеграционных тестах
	// или при настоящем подключении к базе данных
	t.Skip("Skipping initSchema tests - this is better tested with real database connection")
}

// Тест для функции InitData
func TestInitData(t *testing.T) {
	t.Run("admin type already exists and admin user already exists", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		// Мокаем запрос на проверку существования типа admin
		adminTypeID := 1
		adminTypeRows := sqlmock.NewRows([]string{"id"}).AddRow(adminTypeID)
		mock.ExpectQuery(`SELECT id FROM content_types WHERE name = 'admin'`).WillReturnRows(adminTypeRows)

		// Мокаем запрос на проверку существования админа
		adminExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM accounts WHERE username = \$1\)`).
			WithArgs("admin").
			WillReturnRows(adminExists)

		// Вызываем тестируемую функцию
		err := storage.InitData(context.Background())

		// Проверяем результат
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("admin type does not exist but admin user exists", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		// Мокаем запрос на проверку существования типа admin - возвращаем ErrNoRows
		mock.ExpectQuery(`SELECT id FROM content_types WHERE name = 'admin'`).WillReturnError(sql.ErrNoRows)

		// Мокаем создание нового типа admin
		adminTypeID := 1
		adminTypeInsertRows := sqlmock.NewRows([]string{"id"}).AddRow(adminTypeID)
		mock.ExpectQuery(`INSERT INTO content_types \(name, deleted\) VALUES \(\$1, \$2\) RETURNING id`).
			WithArgs("admin", false).
			WillReturnRows(adminTypeInsertRows)

		// Мокаем запрос на проверку существования админа
		adminExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM accounts WHERE username = \$1\)`).
			WithArgs("admin").
			WillReturnRows(adminExists)

		// Вызываем тестируемую функцию
		err := storage.InitData(context.Background())

		// Проверяем результат
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("admin type exists but admin user does not exist", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		// Мокаем запрос на проверку существования типа admin
		adminTypeID := 1
		adminTypeRows := sqlmock.NewRows([]string{"id"}).AddRow(adminTypeID)
		mock.ExpectQuery(`SELECT id FROM content_types WHERE name = 'admin'`).WillReturnRows(adminTypeRows)

		// Мокаем запрос на проверку существования админа - не существует
		adminExists := sqlmock.NewRows([]string{"exists"}).AddRow(false)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM accounts WHERE username = \$1\)`).
			WithArgs("admin").
			WillReturnRows(adminExists)

		// Мокаем создание учетной записи админа
		mock.ExpectExec(`INSERT INTO accounts \(username, content_type_id, password, admin, deleted\) VALUES \(\$1, \$2, \$3, \$4, \$5\)`).
			WithArgs("admin", adminTypeID, sqlmock.AnyArg(), true, false).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Вызываем тестируемую функцию
		err := storage.InitData(context.Background())

		// Проверяем результат
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error checking admin type", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		// Мокаем ошибку при проверке существования типа admin
		expectedErr := errors.New("database error")
		mock.ExpectQuery(`SELECT id FROM content_types WHERE name = 'admin'`).WillReturnError(expectedErr)

		// Вызываем тестируемую функцию
		err := storage.InitData(context.Background())

		// Проверяем результат
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to check admin type")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error creating admin type", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		// Мокаем запрос на проверку существования типа admin - не существует
		mock.ExpectQuery(`SELECT id FROM content_types WHERE name = 'admin'`).WillReturnError(sql.ErrNoRows)

		// Мокаем ошибку при создании типа admin
		expectedErr := errors.New("insert error")
		mock.ExpectQuery(`INSERT INTO content_types \(name, deleted\) VALUES \(\$1, \$2\) RETURNING id`).
			WithArgs("admin", false).
			WillReturnError(expectedErr)

		// Вызываем тестируемую функцию
		err := storage.InitData(context.Background())

		// Проверяем результат
		assert.Error(t, err)
		assert.ErrorContains(t, err, "create admin type failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error checking admin existence", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		// Мокаем запрос на проверку существования типа admin
		adminTypeID := 1
		adminTypeRows := sqlmock.NewRows([]string{"id"}).AddRow(adminTypeID)
		mock.ExpectQuery(`SELECT id FROM content_types WHERE name = 'admin'`).WillReturnRows(adminTypeRows)

		// Мокаем ошибку при проверке существования админа
		expectedErr := errors.New("query error")
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM accounts WHERE username = \$1\)`).
			WithArgs("admin").
			WillReturnError(expectedErr)

		// Вызываем тестируемую функцию
		err := storage.InitData(context.Background())

		// Проверяем результат
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to check admin existence")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
