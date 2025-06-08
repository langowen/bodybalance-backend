package admin

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTypesTestDB вспомогательная функция для настройки тестового окружения
func setupTypesTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Storage) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")

	storage := &Storage{
		db: db,
	}

	return db, mock, storage
}

// TestAddType тестирует функцию AddType
func TestAddType(t *testing.T) {
	t.Run("successful add", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		req := admResponse.TypeRequest{
			Name: "Test Type",
		}

		// Ожидаем INSERT запрос для добавления типа
		createdAt := time.Now()
		rows := sqlmock.NewRows([]string{"id", "name", "created_at"}).
			AddRow(float64(1), req.Name, createdAt)

		mock.ExpectQuery(`INSERT INTO content_types`).
			WithArgs(req.Name).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		typeResponse, err := storage.AddType(ctx, req)

		// Проверка результатов
		require.NoError(t, err)
		assert.Equal(t, float64(1), typeResponse.ID)
		assert.Equal(t, req.Name, typeResponse.Name)
		assert.Equal(t, createdAt.Format("02.01.2006"), typeResponse.DateCreated)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		req := admResponse.TypeRequest{
			Name: "Test Type",
		}

		// Ожидаем ошибку при добавлении типа
		mock.ExpectQuery(`INSERT INTO content_types`).
			WithArgs(req.Name).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		_, err := storage.AddType(ctx, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "database error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetType тестирует функцию GetType
func TestGetType(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		typeID := int64(1)
		createdAt := time.Now()

		// Ожидаем SELECT запрос для получения типа
		rows := sqlmock.NewRows([]string{"id", "name", "created_at"}).
			AddRow(float64(1), "Test Type", createdAt)

		mock.ExpectQuery(`SELECT id, name, created_at FROM content_types`).
			WithArgs(typeID).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		typeResponse, err := storage.GetType(ctx, typeID)

		// Проверка результатов
		require.NoError(t, err)
		assert.Equal(t, float64(1), typeResponse.ID)
		assert.Equal(t, "Test Type", typeResponse.Name)
		assert.Equal(t, createdAt.Format("02.01.2006"), typeResponse.DateCreated)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("type not found", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		typeID := int64(999)

		// Ожидаем SELECT запрос без результатов
		mock.ExpectQuery(`SELECT id, name, created_at FROM content_types`).
			WithArgs(typeID).
			WillReturnError(sql.ErrNoRows)

		// Вызов тестируемого метода
		_, err := storage.GetType(ctx, typeID)

		// Проверка результатов
		require.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		typeID := int64(1)

		// Ожидаем ошибку при запросе
		mock.ExpectQuery(`SELECT id, name, created_at FROM content_types`).
			WithArgs(typeID).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		_, err := storage.GetType(ctx, typeID)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "database error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetTypes тестирует функцию GetTypes
func TestGetTypes(t *testing.T) {
	t.Run("successful get all", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		createdAt1 := time.Now()
		createdAt2 := time.Now().Add(-24 * time.Hour)

		// Ожидаем SELECT запрос для получения всех типов
		rows := sqlmock.NewRows([]string{"id", "name", "created_at"}).
			AddRow(float64(1), "Type 1", createdAt1).
			AddRow(float64(2), "Type 2", createdAt2)

		mock.ExpectQuery(`SELECT id, name, created_at FROM content_types`).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		types, err := storage.GetTypes(ctx)

		// Проверка результатов
		require.NoError(t, err)
		assert.Len(t, types, 2)

		assert.Equal(t, float64(1), types[0].ID)
		assert.Equal(t, "Type 1", types[0].Name)
		assert.Equal(t, createdAt1.Format("02.01.2006"), types[0].DateCreated)

		assert.Equal(t, float64(2), types[1].ID)
		assert.Equal(t, "Type 2", types[1].Name)
		assert.Equal(t, createdAt2.Format("02.01.2006"), types[1].DateCreated)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()

		// Ожидаем SELECT запрос для получения всех типов (пустой результат)
		rows := sqlmock.NewRows([]string{"id", "name", "created_at"})

		mock.ExpectQuery(`SELECT id, name, created_at FROM content_types`).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		types, err := storage.GetTypes(ctx)

		// Проверка результатов
		require.NoError(t, err)
		assert.Empty(t, types)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()

		// Ожидаем ошибку при запросе
		mock.ExpectQuery(`SELECT id, name, created_at FROM content_types`).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		_, err := storage.GetTypes(ctx)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "database error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()

		// Ожидаем SELECT запрос с некорректными данными в результате
		rows := sqlmock.NewRows([]string{"id", "name", "created_at"}).
			AddRow("not-a-float", "Type 1", time.Now()) // ID должен быть числом

		mock.ExpectQuery(`SELECT id, name, created_at FROM content_types`).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		_, err := storage.GetTypes(ctx)

		// Проверка результатов
		require.Error(t, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestUpdateType тестирует функцию UpdateType
func TestUpdateType(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		typeID := int64(1)
		req := admResponse.TypeRequest{
			Name: "Updated Type",
		}

		// Ожидаем UPDATE запрос для обновления типа
		mock.ExpectExec(`UPDATE content_types SET name = \$1 WHERE id = \$2`).
			WithArgs(req.Name, typeID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Вызов тестируемого метода
		err := storage.UpdateType(ctx, typeID, req)

		// Проверка результатов
		require.NoError(t, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("type not found", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		typeID := int64(999)
		req := admResponse.TypeRequest{
			Name: "Updated Type",
		}

		// Ожидаем UPDATE запрос, который не затрагивает ни одной строки
		mock.ExpectExec(`UPDATE content_types SET name = \$1 WHERE id = \$2`).
			WithArgs(req.Name, typeID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Вызов тестируемого метода
		err := storage.UpdateType(ctx, typeID, req)

		// Проверка результатов
		require.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		typeID := int64(1)
		req := admResponse.TypeRequest{
			Name: "Updated Type",
		}

		// Ожидаем ошибку при выполнении UPDATE запроса
		mock.ExpectExec(`UPDATE content_types SET name = \$1 WHERE id = \$2`).
			WithArgs(req.Name, typeID).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		err := storage.UpdateType(ctx, typeID, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "database error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		typeID := int64(1)
		req := admResponse.TypeRequest{
			Name: "Updated Type",
		}

		// Ожидаем UPDATE запрос с ошибкой при проверке затронутых строк
		result := sqlmock.NewErrorResult(errors.New("rows affected error"))
		mock.ExpectExec(`UPDATE content_types SET name = \$1 WHERE id = \$2`).
			WithArgs(req.Name, typeID).
			WillReturnResult(result)

		// Вызов тестируемого метода
		err := storage.UpdateType(ctx, typeID, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "rows affected error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestDeleteType тестирует функцию DeleteType
func TestDeleteType(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		typeID := int64(1)

		// Ожидаем UPDATE запрос для пометки типа как удаленного
		mock.ExpectExec(`UPDATE content_types SET deleted = TRUE WHERE id = \$1`).
			WithArgs(typeID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Вызов тестируемого метода
		err := storage.DeleteType(ctx, typeID)

		// Проверка результатов
		require.NoError(t, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("type not found", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		typeID := int64(999)

		// Ожидаем UPDATE запрос, который не затрагивает ни одной строки
		mock.ExpectExec(`UPDATE content_types SET deleted = TRUE WHERE id = \$1`).
			WithArgs(typeID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Вызов тестируемого метода
		err := storage.DeleteType(ctx, typeID)

		// Проверка результатов
		require.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		typeID := int64(1)

		// Ожидаем ошибку при выполнении UPDATE запроса
		mock.ExpectExec(`UPDATE content_types SET deleted = TRUE WHERE id = \$1`).
			WithArgs(typeID).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		err := storage.DeleteType(ctx, typeID)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "database error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupTypesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		typeID := int64(1)

		// Ожидаем UPDATE запрос с ошибкой при проверке затронутых строк
		result := sqlmock.NewErrorResult(errors.New("rows affected error"))
		mock.ExpectExec(`UPDATE content_types SET deleted = TRUE WHERE id = \$1`).
			WithArgs(typeID).
			WillReturnResult(result)

		// Вызов тестируемого метода
		err := storage.DeleteType(ctx, typeID)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "rows affected error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
