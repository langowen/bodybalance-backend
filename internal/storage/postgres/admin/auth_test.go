package admin

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Вспомогательная функция для создания тестового окружения
func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Storage) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")

	// Создаем экземпляр хранилища с моком DB
	storage := &Storage{
		db: db,
	}

	return db, mock, storage
}

// TestGetAdminUser тестирует функцию GetAdminUser
func TestGetAdminUser(t *testing.T) {
	t.Run("successful authentication", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		login := "admin"
		passwordHash := "hashed_password"

		// Настройка ожидаемого поведения
		rows := sqlmock.NewRows([]string{"username", "password", "admin"}).
			AddRow(login, passwordHash, true)

		mock.ExpectQuery(`SELECT username, password, admin.*FROM accounts.*WHERE username = \$1 AND password = \$2`).
			WithArgs(login, passwordHash).
			WillReturnRows(rows)

		// Вызов тестируемой функции
		user, err := storage.GetAdminUser(ctx, login, passwordHash)

		// Проверка результатов
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, login, user.Username)
		assert.Equal(t, passwordHash, user.Password)
		assert.True(t, user.IsAdmin)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		login := "nonexistent"
		passwordHash := "wrong_password"

		// Настройка ожидаемого поведения - пользователь не найден
		mock.ExpectQuery(`SELECT username, password, admin.*FROM accounts.*WHERE username = \$1 AND password = \$2`).
			WithArgs(login, passwordHash).
			WillReturnError(sql.ErrNoRows)

		// Вызов тестируемой функции
		user, err := storage.GetAdminUser(ctx, login, passwordHash)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		login := "admin"
		passwordHash := "hashed_password"

		// Настройка ожидаемого поведения - ошибка базы данных
		dbError := errors.New("database connection error")
		mock.ExpectQuery(`SELECT username, password, admin.*FROM accounts.*WHERE username = \$1 AND password = \$2`).
			WithArgs(login, passwordHash).
			WillReturnError(dbError)

		// Вызов тестируемой функции
		user, err := storage.GetAdminUser(ctx, login, passwordHash)

		// Проверка результатов
		require.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorContains(t, err, "database connection error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		login := "admin"
		passwordHash := "hashed_password"

		// Настройка ожидаемого поведения - возвращаем неправильные данные для проверки ошибки сканирования
		rows := sqlmock.NewRows([]string{"username"}).
			AddRow(login) // Недостаточно столбцов для сканирования

		mock.ExpectQuery(`SELECT username, password, admin.*FROM accounts.*WHERE username = \$1 AND password = \$2`).
			WithArgs(login, passwordHash).
			WillReturnRows(rows)

		// Вызов тестируемой функции
		user, err := storage.GetAdminUser(ctx, login, passwordHash)

		// Проверка результатов
		require.Error(t, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
