package admin

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupUsersTestDB вспомогательная функция для настройки тестового окружения
func setupUsersTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Storage) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")

	storage := &Storage{
		db: db,
	}

	return db, mock, storage
}

// TestAddUser тестирует функцию AddUser
func TestAddUser(t *testing.T) {
	t.Run("successful add", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		req := admResponse.UserRequest{
			Username:      "testuser",
			ContentTypeID: "1",
			Admin:         false,
			Password:      "password123",
		}

		// Ожидаем INSERT запрос для добавления пользователя
		createdAt := time.Now()
		rows := sqlmock.NewRows([]string{"id", "username", "content_type_id", "content_type", "admin", "created_at"}).
			AddRow("1", req.Username, req.ContentTypeID, "TestType", req.Admin, createdAt)

		mock.ExpectQuery(`INSERT INTO accounts \(username, content_type_id, admin, password, deleted\) VALUES \(\$1, \$2, \$3, \$4, FALSE\) RETURNING id, username, content_type_id, \(SELECT name FROM content_types WHERE id = \$2\), admin, created_at`).
			WithArgs(req.Username, req.ContentTypeID, req.Admin, req.Password).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		user, err := storage.AddUser(ctx, req)

		// Проверка результатов
		require.NoError(t, err)
		assert.Equal(t, "1", user.ID)
		assert.Equal(t, req.Username, user.Username)
		assert.Equal(t, req.ContentTypeID, user.ContentTypeID)
		assert.Equal(t, "TestType", user.ContentType)
		assert.Equal(t, req.Admin, user.Admin)
		assert.Equal(t, createdAt.Format("02.01.2006"), user.DateCreated)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate user", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		req := admResponse.UserRequest{
			Username:      "existinguser",
			ContentTypeID: "1",
			Admin:         false,
			Password:      "password123",
		}

		// Ожидаем ошибку дубликата при добавлении пользователя
		duplicateErr := errors.New("duplicate key value violates unique constraint")
		mock.ExpectQuery(`INSERT INTO accounts`).
			WithArgs(req.Username, req.ContentTypeID, req.Admin, req.Password).
			WillReturnError(duplicateErr)

		// Вызов тестируемого метода
		_, err := storage.AddUser(ctx, req)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "user already exists"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		req := admResponse.UserRequest{
			Username:      "testuser",
			ContentTypeID: "1",
			Admin:         false,
			Password:      "password123",
		}

		// Ожидаем ошибку при добавлении пользователя
		dbErr := errors.New("database error")
		mock.ExpectQuery(`INSERT INTO accounts`).
			WithArgs(req.Username, req.ContentTypeID, req.Admin, req.Password).
			WillReturnError(dbErr)

		// Вызов тестируемого метода
		_, err := storage.AddUser(ctx, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "database error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetUser тестирует функцию GetUser
func TestGetUser(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		userID := int64(1)
		createdAt := time.Now()

		// Ожидаем SELECT запрос для получения пользователя
		rows := sqlmock.NewRows([]string{"id", "username", "content_type_id", "content_type", "admin", "created_at"}).
			AddRow(float64(1), "testuser", 1, "TestType", false, createdAt)

		mock.ExpectQuery(`SELECT a\.id, a\.username, a\.content_type_id, ct\.name, a\.admin, a\.created_at FROM accounts a LEFT JOIN content_types ct ON a\.content_type_id = ct\.id WHERE a\.id = \$1 AND a\.deleted IS NOT TRUE`).
			WithArgs(userID).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		user, err := storage.GetUser(ctx, userID)

		// Проверка результатов
		require.NoError(t, err)
		assert.Equal(t, "1", user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "1", user.ContentTypeID)
		assert.Equal(t, "TestType", user.ContentType)
		assert.Equal(t, false, user.Admin)
		assert.Equal(t, createdAt.Format("02.01.2006"), user.DateCreated)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		userID := int64(999)

		// Ожидаем SELECT запрос для несуществующего пользователя
		mock.ExpectQuery(`SELECT a\.id, a\.username, a\.content_type_id, ct\.name, a\.admin, a\.created_at FROM accounts a LEFT JOIN content_types ct ON a\.content_type_id = ct\.id WHERE a\.id = \$1 AND a\.deleted IS NOT TRUE`).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		// Вызов тестируемого метода
		_, err := storage.GetUser(ctx, userID)

		// Проверка результатов
		require.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		userID := int64(1)

		// Ожидаем ошибку базы данных при запросе
		mock.ExpectQuery(`SELECT a\.id, a\.username, a\.content_type_id, ct\.name, a\.admin, a\.created_at FROM accounts a LEFT JOIN content_types ct ON a\.content_type_id = ct\.id WHERE a\.id = \$1 AND a\.deleted IS NOT TRUE`).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		_, err := storage.GetUser(ctx, userID)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "database error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetUsers тестирует функцию GetUsers
func TestGetUsers(t *testing.T) {
	t.Run("successful get all", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		createdAt1 := time.Now()
		createdAt2 := time.Now().Add(-24 * time.Hour)

		// Ожидаем SELECT запрос для получения всех пользователей
		rows := sqlmock.NewRows([]string{"id", "username", "content_type_id", "content_type", "admin", "created_at"}).
			AddRow(float64(1), "user1", 1, "Type1", false, createdAt1).
			AddRow(float64(2), "user2", 2, "Type2", true, createdAt2)

		mock.ExpectQuery(`SELECT a\.id, a\.username, a\.content_type_id, ct\.name, a\.admin, a\.created_at FROM accounts a LEFT JOIN content_types ct ON a\.content_type_id = ct\.id WHERE a\.deleted IS NOT TRUE ORDER BY a\.id`).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		users, err := storage.GetUsers(ctx)

		// Проверка результатов
		require.NoError(t, err)
		assert.Len(t, users, 2)

		// Проверка первого пользователя
		assert.Equal(t, "1", users[0].ID)
		assert.Equal(t, "user1", users[0].Username)
		assert.Equal(t, "1", users[0].ContentTypeID)
		assert.Equal(t, "Type1", users[0].ContentType)
		assert.Equal(t, false, users[0].Admin)
		assert.Equal(t, createdAt1.Format("02.01.2006"), users[0].DateCreated)

		// Проверка второго пользователя
		assert.Equal(t, "2", users[1].ID)
		assert.Equal(t, "user2", users[1].Username)
		assert.Equal(t, "2", users[1].ContentTypeID)
		assert.Equal(t, "Type2", users[1].ContentType)
		assert.Equal(t, true, users[1].Admin)
		assert.Equal(t, createdAt2.Format("02.01.2006"), users[1].DateCreated)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()

		// Ожидаем SELECT запрос для получения пустого списка пользователей
		rows := sqlmock.NewRows([]string{"id", "username", "content_type_id", "content_type", "admin", "created_at"})

		mock.ExpectQuery(`SELECT a\.id, a\.username, a\.content_type_id, ct\.name, a\.admin, a\.created_at FROM accounts a LEFT JOIN content_types ct ON a\.content_type_id = ct\.id WHERE a\.deleted IS NOT TRUE ORDER BY a\.id`).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		users, err := storage.GetUsers(ctx)

		// Проверка результатов
		require.NoError(t, err)
		assert.Empty(t, users)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()

		// Ожидаем ошибку базы данных при запросе
		mock.ExpectQuery(`SELECT a\.id, a\.username, a\.content_type_id, ct\.name, a\.admin, a\.created_at FROM accounts a LEFT JOIN content_types ct ON a\.content_type_id = ct\.id WHERE a\.deleted IS NOT TRUE ORDER BY a\.id`).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		_, err := storage.GetUsers(ctx)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "database error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()

		// Создаем строки с некорректными данными, что гарантировано вызовет ошибку сканирования
		rows := sqlmock.NewRows([]string{"id", "username", "content_type_id", "content_type", "admin", "created_at"}).
			AddRow(nil, "user1", "1", "Type1", false, time.Now()) // nil для поля ID вызовет ошибку

		mock.ExpectQuery(`SELECT a\.id, a\.username, a\.content_type_id, ct\.name, a\.admin, a\.created_at FROM accounts a LEFT JOIN content_types ct ON a\.content_type_id = ct\.id WHERE a\.deleted IS NOT TRUE ORDER BY a\.id`).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		_, err := storage.GetUsers(ctx)

		// Проверка результатов
		require.Error(t, err)

		// Проверяем, что ошибка содержит упоминание о сканировании, используя регистронезависимую проверку
		errMsg := strings.ToLower(err.Error())
		assert.Contains(t, errMsg, "scan")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestUpdateUser тестирует функцию UpdateUser
func TestUpdateUser(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		userID := int64(1)
		req := admResponse.UserRequest{
			Username:      "updateduser",
			ContentTypeID: "2",
			Admin:         true,
			Password:      "newpassword",
		}

		// Ожидаем UPDATE запрос для обновления пользователя
		mock.ExpectExec(`UPDATE accounts SET username = \$1, content_type_id = \$2, admin = \$3, password = \$4 WHERE id = \$5 AND deleted IS NOT TRUE`).
			WithArgs(req.Username, req.ContentTypeID, req.Admin, req.Password, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Вызов тестируемого метода
		err := storage.UpdateUser(ctx, userID, req)

		// Проверка результатов
		require.NoError(t, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		userID := int64(999)
		req := admResponse.UserRequest{
			Username:      "updateduser",
			ContentTypeID: "2",
			Admin:         true,
			Password:      "newpassword",
		}

		// Ожидаем UPDATE запрос для несуществующего пользователя (0 затронутых строк)
		mock.ExpectExec(`UPDATE accounts SET username = \$1, content_type_id = \$2, admin = \$3, password = \$4 WHERE id = \$5 AND deleted IS NOT TRUE`).
			WithArgs(req.Username, req.ContentTypeID, req.Admin, req.Password, userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Вызов тестируемого метода
		err := storage.UpdateUser(ctx, userID, req)

		// Проверка результатов
		require.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		userID := int64(1)
		req := admResponse.UserRequest{
			Username:      "updateduser",
			ContentTypeID: "2",
			Admin:         true,
			Password:      "newpassword",
		}

		// Ожидаем ошибку базы данных при обновлении
		mock.ExpectExec(`UPDATE accounts SET username = \$1, content_type_id = \$2, admin = \$3, password = \$4 WHERE id = \$5 AND deleted IS NOT TRUE`).
			WithArgs(req.Username, req.ContentTypeID, req.Admin, req.Password, userID).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		err := storage.UpdateUser(ctx, userID, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "database error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		userID := int64(1)
		req := admResponse.UserRequest{
			Username:      "updateduser",
			ContentTypeID: "2",
			Admin:         true,
			Password:      "newpassword",
		}

		// Ожидаем UPDATE запрос с ошибкой при получении затронутых строк
		result := sqlmock.NewErrorResult(errors.New("rows affected error"))
		mock.ExpectExec(`UPDATE accounts SET username = \$1, content_type_id = \$2, admin = \$3, password = \$4 WHERE id = \$5 AND deleted IS NOT TRUE`).
			WithArgs(req.Username, req.ContentTypeID, req.Admin, req.Password, userID).
			WillReturnResult(result)

		// Вызов тестируемого метода
		err := storage.UpdateUser(ctx, userID, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "rows affected error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestDeleteUser тестирует функцию DeleteUser
func TestDeleteUser(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		userID := int64(1)

		// Ожидаем UPDATE запрос для пометки пользователя как удаленного
		mock.ExpectExec(`UPDATE accounts SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Вызов тестируемого метода
		err := storage.DeleteUser(ctx, userID)

		// Проверка результатов
		require.NoError(t, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		userID := int64(999)

		// Ожидаем UPDATE запрос для несуществующего пользователя (0 затронутых строк)
		mock.ExpectExec(`UPDATE accounts SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Вызов тестируемого метода
		err := storage.DeleteUser(ctx, userID)

		// Проверка результатов
		require.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		userID := int64(1)

		// Ожидаем ошибку базы данных при удалении
		mock.ExpectExec(`UPDATE accounts SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		err := storage.DeleteUser(ctx, userID)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "database error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupUsersTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		userID := int64(1)

		// Ожидаем UPDATE запрос с ошибкой при получении затронутых строк
		result := sqlmock.NewErrorResult(errors.New("rows affected error"))
		mock.ExpectExec(`UPDATE accounts SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
			WithArgs(userID).
			WillReturnResult(result)

		// Вызов тестируемого метода
		err := storage.DeleteUser(ctx, userID)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "rows affected error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
