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

// Вспомогательная функция для настройки тестового окружения
func setupCategoriesTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Storage) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")

	storage := &Storage{
		db: db,
	}

	return db, mock, storage
}

// TestAddCategory тестирует функцию AddCategory
func TestAddCategory(t *testing.T) {
	t.Run("successful add", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		req := admResponse.CategoryRequest{
			Name:    "Test Category",
			ImgURL:  "test.jpg",
			TypeIDs: []int64{1, 2},
		}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем INSERT запрос для добавления категории
		createdAt := time.Now()
		categoryRows := sqlmock.NewRows([]string{"id", "name", "img_url", "created_at"}).
			AddRow(float64(1), req.Name, req.ImgURL, createdAt)

		mock.ExpectQuery(`INSERT INTO categories`).
			WithArgs(req.Name, req.ImgURL).
			WillReturnRows(categoryRows)

		// Ожидаем INSERT запросы для добавления связей с типами
		for _, typeID := range req.TypeIDs {
			mock.ExpectExec(`INSERT INTO category_content_types`).
				WithArgs(float64(1), typeID).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}

		// Ожидаем запрос на получение связанных типов
		typeRows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(float64(1), "Type 1").
			AddRow(float64(2), "Type 2")

		mock.ExpectQuery(`SELECT ct\.id, ct\.name FROM content_types ct JOIN category_content_types`).
			WithArgs(float64(1)).
			WillReturnRows(typeRows)

		// Ожидаем коммит транзакции
		mock.ExpectCommit()

		// Вызов тестируемого метода
		category, err := storage.AddCategory(ctx, req)

		// Проверка результатов
		require.NoError(t, err)
		assert.Equal(t, float64(1), category.ID)
		assert.Equal(t, req.Name, category.Name)
		assert.Equal(t, req.ImgURL, category.ImgURL)
		assert.Equal(t, createdAt.Format("02.01.2006"), category.DateCreated)
		assert.Len(t, category.Types, 2)
		assert.Equal(t, float64(1), category.Types[0].ID)
		assert.Equal(t, "Type 1", category.Types[0].Name)
		assert.Equal(t, float64(2), category.Types[1].ID)
		assert.Equal(t, "Type 2", category.Types[1].Name)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("transaction begin error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		req := admResponse.CategoryRequest{
			Name:    "Test Category",
			ImgURL:  "test.jpg",
			TypeIDs: []int64{1, 2},
		}

		// Ожидаем ошибку при начале транзакции
		mock.ExpectBegin().WillReturnError(errors.New("begin error"))

		// Вызов тестируемого метода
		_, err := storage.AddCategory(ctx, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "begin error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("insert category error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		req := admResponse.CategoryRequest{
			Name:    "Test Category",
			ImgURL:  "test.jpg",
			TypeIDs: []int64{1, 2},
		}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем ошибку при добавлении категории
		mock.ExpectQuery(`INSERT INTO categories`).
			WithArgs(req.Name, req.ImgURL).
			WillReturnError(errors.New("insert error"))

		// Ожидаем откат транзакции из-за ошибки
		mock.ExpectRollback()

		// Вызов тестируемого метода
		_, err := storage.AddCategory(ctx, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "insert error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("insert type relation error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		req := admResponse.CategoryRequest{
			Name:    "Test Category",
			ImgURL:  "test.jpg",
			TypeIDs: []int64{1, 2},
		}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем INSERT запрос для добавления категории
		createdAt := time.Now()
		categoryRows := sqlmock.NewRows([]string{"id", "name", "img_url", "created_at"}).
			AddRow(float64(1), req.Name, req.ImgURL, createdAt)

		mock.ExpectQuery(`INSERT INTO categories`).
			WithArgs(req.Name, req.ImgURL).
			WillReturnRows(categoryRows)

		// Ожидаем ошибку при добавлении связи с типом
		mock.ExpectExec(`INSERT INTO category_content_types`).
			WithArgs(float64(1), req.TypeIDs[0]).
			WillReturnError(errors.New("relation insert error"))

		// Ожидаем откат транзакции из-за ошибки
		mock.ExpectRollback()

		// Вызов тестируемого метода
		_, err := storage.AddCategory(ctx, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "relation insert error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("select types error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		req := admResponse.CategoryRequest{
			Name:    "Test Category",
			ImgURL:  "test.jpg",
			TypeIDs: []int64{1, 2},
		}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем INSERT запрос для добавления категории
		createdAt := time.Now()
		categoryRows := sqlmock.NewRows([]string{"id", "name", "img_url", "created_at"}).
			AddRow(float64(1), req.Name, req.ImgURL, createdAt)

		mock.ExpectQuery(`INSERT INTO categories`).
			WithArgs(req.Name, req.ImgURL).
			WillReturnRows(categoryRows)

		// Ожидаем INSERT запросы для добавления связей с типами
		for _, typeID := range req.TypeIDs {
			mock.ExpectExec(`INSERT INTO category_content_types`).
				WithArgs(float64(1), typeID).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}

		// Ожидаем ошибку при запросе типов
		mock.ExpectQuery(`SELECT ct\.id, ct\.name FROM content_types ct JOIN category_content_types`).
			WithArgs(float64(1)).
			WillReturnError(errors.New("select types error"))

		// Ожидаем откат транзакции из-за ошибки
		mock.ExpectRollback()

		// Вызов тестируемого метода
		_, err := storage.AddCategory(ctx, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "select types error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("commit error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		req := admResponse.CategoryRequest{
			Name:    "Test Category",
			ImgURL:  "test.jpg",
			TypeIDs: []int64{1, 2},
		}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем INSERT запрос для добавления категории
		createdAt := time.Now()
		categoryRows := sqlmock.NewRows([]string{"id", "name", "img_url", "created_at"}).
			AddRow(float64(1), req.Name, req.ImgURL, createdAt)

		mock.ExpectQuery(`INSERT INTO categories`).
			WithArgs(req.Name, req.ImgURL).
			WillReturnRows(categoryRows)

		// Ожидаем INSERT запросы для добавления связей с типами
		for _, typeID := range req.TypeIDs {
			mock.ExpectExec(`INSERT INTO category_content_types`).
				WithArgs(float64(1), typeID).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}

		// Ожидаем запрос на получение связанных типов
		typeRows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(float64(1), "Type 1").
			AddRow(float64(2), "Type 2")

		mock.ExpectQuery(`SELECT ct\.id, ct\.name FROM content_types ct JOIN category_content_types`).
			WithArgs(float64(1)).
			WillReturnRows(typeRows)

		// Ожидаем ошибку при коммите транзакции
		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		// Вызов тестируемого метода
		_, err := storage.AddCategory(ctx, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "commit error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetCategory тестирует функцию GetCategory
func TestGetCategory(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(1)
		createdAt := time.Now()

		// Ожидаем запрос для получения категории
		categoryRows := sqlmock.NewRows([]string{"id", "name", "img_url", "created_at"}).
			AddRow(float64(1), "Test Category", "test.jpg", createdAt)

		mock.ExpectQuery(`SELECT id, name, img_url, created_at FROM categories WHERE id = \$1`).
			WithArgs(categoryID).
			WillReturnRows(categoryRows)

		// Ожидаем запрос на получение связанных типов
		typeRows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(float64(1), "Type 1").
			AddRow(float64(2), "Type 2")

		mock.ExpectQuery(`SELECT ct\.id, ct\.name FROM content_types ct JOIN category_content_types`).
			WithArgs(categoryID). // Используем categoryID типа int64 вместо float64(1)
			WillReturnRows(typeRows)

		// Вызов тестируемого метода
		category, err := storage.GetCategory(ctx, categoryID)

		// Проверка результатов
		require.NoError(t, err)
		assert.Equal(t, float64(1), category.ID)
		assert.Equal(t, "Test Category", category.Name)
		assert.Equal(t, "test.jpg", category.ImgURL)
		assert.Equal(t, createdAt.Format("02.01.2006"), category.DateCreated)
		assert.Len(t, category.Types, 2)
		assert.Equal(t, float64(1), category.Types[0].ID)
		assert.Equal(t, "Type 1", category.Types[0].Name)
		assert.Equal(t, float64(2), category.Types[1].ID)
		assert.Equal(t, "Type 2", category.Types[1].Name)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("category not found", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(999)

		// Ожидаем запрос для получения категории, которая не существует
		mock.ExpectQuery(`SELECT id, name, img_url, created_at FROM categories WHERE id = \$1`).
			WithArgs(categoryID).
			WillReturnError(sql.ErrNoRows)

		// Вызов тестируемого метода
		_, err := storage.GetCategory(ctx, categoryID)

		// Проверка результатов
		require.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(1)

		// Ожидаем запрос с ошибкой базы данных
		mock.ExpectQuery(`SELECT id, name, img_url, created_at FROM categories WHERE id = \$1`).
			WithArgs(categoryID).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		_, err := storage.GetCategory(ctx, categoryID)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "database error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error getting types", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(1)
		createdAt := time.Now()

		// Ожидаем запрос для получения категории
		categoryRows := sqlmock.NewRows([]string{"id", "name", "img_url", "created_at"}).
			AddRow(float64(1), "Test Category", "test.jpg", createdAt)

		mock.ExpectQuery(`SELECT id, name, img_url, created_at FROM categories WHERE id = \$1`).
			WithArgs(categoryID).
			WillReturnRows(categoryRows)

		// Ожидаем запрос на получение типов с ошибкой
		mock.ExpectQuery(`SELECT ct\.id, ct\.name FROM content_types ct JOIN category_content_types`).
			WithArgs(categoryID). // Используем categoryID вместо float64(1)
			WillReturnError(errors.New("types query error"))

		// Вызов тестируемого метода
		_, err := storage.GetCategory(ctx, categoryID)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "types query error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetCategories тестирует функцию GetCategories
func TestGetCategories(t *testing.T) {
	t.Run("successful get all", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		createdAt1 := time.Now()
		createdAt2 := time.Now().Add(-24 * time.Hour)

		// Ожидаем запрос для получения всех категорий
		categoryRows := sqlmock.NewRows([]string{"id", "name", "img_url", "created_at"}).
			AddRow(float64(1), "Category 1", "img1.jpg", createdAt1).
			AddRow(float64(2), "Category 2", "img2.jpg", createdAt2)

		mock.ExpectQuery(`SELECT id, name, img_url, created_at FROM categories`).
			WillReturnRows(categoryRows)

		// Ожидаем запрос на получение типов для первой категории
		type1Rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(float64(1), "Type 1").
			AddRow(float64(2), "Type 2")

		mock.ExpectQuery(`SELECT ct\.id, ct\.name FROM content_types ct JOIN category_content_types`).
			WithArgs(float64(1)).
			WillReturnRows(type1Rows)

		// Ожидаем запрос на получение типов для второй категории
		type2Rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(float64(2), "Type 2").
			AddRow(float64(3), "Type 3")

		mock.ExpectQuery(`SELECT ct\.id, ct\.name FROM content_types ct JOIN category_content_types`).
			WithArgs(float64(2)).
			WillReturnRows(type2Rows)

		// Вызов тестируемого метода
		categories, err := storage.GetCategories(ctx)

		// Проверка результатов
		require.NoError(t, err)
		assert.Len(t, categories, 2)

		// Проверка первой категории
		assert.Equal(t, float64(1), categories[0].ID)
		assert.Equal(t, "Category 1", categories[0].Name)
		assert.Equal(t, "img1.jpg", categories[0].ImgURL)
		assert.Equal(t, createdAt1.Format("02.01.2006"), categories[0].DateCreated)
		assert.Len(t, categories[0].Types, 2)
		assert.Equal(t, float64(1), categories[0].Types[0].ID)
		assert.Equal(t, "Type 1", categories[0].Types[0].Name)
		assert.Equal(t, float64(2), categories[0].Types[1].ID)
		assert.Equal(t, "Type 2", categories[0].Types[1].Name)

		// Проверка второй категории
		assert.Equal(t, float64(2), categories[1].ID)
		assert.Equal(t, "Category 2", categories[1].Name)
		assert.Equal(t, "img2.jpg", categories[1].ImgURL)
		assert.Equal(t, createdAt2.Format("02.01.2006"), categories[1].DateCreated)
		assert.Len(t, categories[1].Types, 2)
		assert.Equal(t, float64(2), categories[1].Types[0].ID)
		assert.Equal(t, "Type 2", categories[1].Types[0].Name)
		assert.Equal(t, float64(3), categories[1].Types[1].ID)
		assert.Equal(t, "Type 3", categories[1].Types[1].Name)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()

		// Ожидаем запрос с ошибкой базы данных
		mock.ExpectQuery(`SELECT id, name, img_url, created_at FROM categories`).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		_, err := storage.GetCategories(ctx)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "database error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error getting types", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		createdAt := time.Now()

		// Ожидаем запрос для получения всех категорий
		categoryRows := sqlmock.NewRows([]string{"id", "name", "img_url", "created_at"}).
			AddRow(float64(1), "Category 1", "img1.jpg", createdAt)

		mock.ExpectQuery(`SELECT id, name, img_url, created_at FROM categories`).
			WillReturnRows(categoryRows)

		// Ожидаем запрос на получение типов с ошибкой
		mock.ExpectQuery(`SELECT ct\.id, ct\.name FROM content_types ct JOIN category_content_types`).
			WithArgs(float64(1)).
			WillReturnError(errors.New("types query error"))

		// Вызов тестируемого метода
		_, err := storage.GetCategories(ctx)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "types query error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestUpdateCategory тестирует функцию UpdateCategory
func TestUpdateCategory(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(1)
		req := admResponse.CategoryRequest{
			Name:    "Updated Category",
			ImgURL:  "updated.jpg",
			TypeIDs: []int64{2, 3},
		}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем UPDATE запрос для обновления категории
		mock.ExpectExec(`UPDATE categories SET name = \$1, img_url = \$2 WHERE id = \$3`).
			WithArgs(req.Name, req.ImgURL, categoryID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Ожидаем DELETE запрос для удаления старых связей с типами
		mock.ExpectExec(`DELETE FROM category_content_types WHERE category_id = \$1`).
			WithArgs(categoryID).
			WillReturnResult(sqlmock.NewResult(0, 2))

		// Ожидаем INSERT запросы для добавления новых связей с типами
		for _, typeID := range req.TypeIDs {
			mock.ExpectExec(`INSERT INTO category_content_types \(category_id, content_type_id\) VALUES \(\$1, \$2\)`).
				WithArgs(categoryID, typeID).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}

		// Ожидаем коммит транзакции
		mock.ExpectCommit()

		// Вызов тестируемого метода
		err := storage.UpdateCategory(ctx, categoryID, req)

		// Проверка результатов
		require.NoError(t, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("transaction begin error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(1)
		req := admResponse.CategoryRequest{
			Name:    "Updated Category",
			ImgURL:  "updated.jpg",
			TypeIDs: []int64{2, 3},
		}

		// Ожидаем ошибку при начале транзакции
		mock.ExpectBegin().WillReturnError(errors.New("begin error"))

		// Вызов тестируемого метода
		err := storage.UpdateCategory(ctx, categoryID, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "begin error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(1)
		req := admResponse.CategoryRequest{
			Name:    "Updated Category",
			ImgURL:  "updated.jpg",
			TypeIDs: []int64{2, 3},
		}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем ошибку при обновлении категории
		mock.ExpectExec(`UPDATE categories SET name = \$1, img_url = \$2 WHERE id = \$3`).
			WithArgs(req.Name, req.ImgURL, categoryID).
			WillReturnError(errors.New("update error"))

		// Ожидаем откат транзакции из-за ошибки
		mock.ExpectRollback()

		// Вызов тестируемого метода
		err := storage.UpdateCategory(ctx, categoryID, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "update error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("category not found", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(999)
		req := admResponse.CategoryRequest{
			Name:    "Updated Category",
			ImgURL:  "updated.jpg",
			TypeIDs: []int64{2, 3},
		}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем UPDATE запрос, который не затрагивает ни одной строки
		mock.ExpectExec(`UPDATE categories SET name = \$1, img_url = \$2 WHERE id = \$3`).
			WithArgs(req.Name, req.ImgURL, categoryID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Ожидаем откат транзакции из-за ошибки
		mock.ExpectRollback()

		// Вызов тестируемого метода
		err := storage.UpdateCategory(ctx, categoryID, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows) // Проверяем, что ошибка содержит sql.ErrNoRows

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete relations error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(1)
		req := admResponse.CategoryRequest{
			Name:    "Updated Category",
			ImgURL:  "updated.jpg",
			TypeIDs: []int64{2, 3},
		}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем UPDATE запрос для обновления категории
		mock.ExpectExec(`UPDATE categories SET name = \$1, img_url = \$2 WHERE id = \$3`).
			WithArgs(req.Name, req.ImgURL, categoryID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Ожидаем ошибку при удалении старых связей
		mock.ExpectExec(`DELETE FROM category_content_types WHERE category_id = \$1`).
			WithArgs(categoryID).
			WillReturnError(errors.New("delete relations error"))

		// Ожидаем откат транзакции из-за ошибки
		mock.ExpectRollback()

		// Вызов тестируемого метода
		err := storage.UpdateCategory(ctx, categoryID, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "delete relations error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("insert relation error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(1)
		req := admResponse.CategoryRequest{
			Name:    "Updated Category",
			ImgURL:  "updated.jpg",
			TypeIDs: []int64{2, 3},
		}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем UPDATE запрос для обновления категории
		mock.ExpectExec(`UPDATE categories SET name = \$1, img_url = \$2 WHERE id = \$3`).
			WithArgs(req.Name, req.ImgURL, categoryID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Ожидаем DELETE запрос для удаления старых связей с типами
		mock.ExpectExec(`DELETE FROM category_content_types WHERE category_id = \$1`).
			WithArgs(categoryID).
			WillReturnResult(sqlmock.NewResult(0, 2))

		// Ожидаем ошибку при добавлении новой связи
		mock.ExpectExec(`INSERT INTO category_content_types \(category_id, content_type_id\) VALUES \(\$1, \$2\)`).
			WithArgs(categoryID, req.TypeIDs[0]).
			WillReturnError(errors.New("insert relation error"))

		// Ожидаем откат транзакции из-за ошибки
		mock.ExpectRollback()

		// Вызов тестируемого метода
		err := storage.UpdateCategory(ctx, categoryID, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "insert relation error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("commit error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(1)
		req := admResponse.CategoryRequest{
			Name:    "Updated Category",
			ImgURL:  "updated.jpg",
			TypeIDs: []int64{2, 3},
		}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем UPDATE запрос для обновления категории
		mock.ExpectExec(`UPDATE categories SET name = \$1, img_url = \$2 WHERE id = \$3`).
			WithArgs(req.Name, req.ImgURL, categoryID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Ожидаем DELETE запрос для удаления старых связей с типами
		mock.ExpectExec(`DELETE FROM category_content_types WHERE category_id = \$1`).
			WithArgs(categoryID).
			WillReturnResult(sqlmock.NewResult(0, 2))

		// Ожидаем INSERT запросы для добавления новых связей с типами
		for _, typeID := range req.TypeIDs {
			mock.ExpectExec(`INSERT INTO category_content_types \(category_id, content_type_id\) VALUES \(\$1, \$2\)`).
				WithArgs(categoryID, typeID).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}

		// Ожидаем ошибку при коммите транзакции
		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		// Вызов тестируемого метода
		err := storage.UpdateCategory(ctx, categoryID, req)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "commit error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestDeleteCategory тестирует функцию DeleteCategory
func TestDeleteCategory(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(1)

		// Ожидаем UPDATE запрос для пометки категории как удаленной
		mock.ExpectExec(`UPDATE categories SET deleted = TRUE WHERE id = \$1`).
			WithArgs(categoryID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Вызов тестируемого метода
		err := storage.DeleteCategory(ctx, categoryID)

		// Проверка результатов
		require.NoError(t, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("category not found", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(999)

		// Ожидаем UPDATE запрос, который не затрагивает ни одной строки
		mock.ExpectExec(`UPDATE categories SET deleted = TRUE WHERE id = \$1`).
			WithArgs(categoryID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Вызов тестируемого метода
		err := storage.DeleteCategory(ctx, categoryID)

		// Проверка результатов
		require.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupCategoriesTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		categoryID := int64(1)

		// Ожидаем ошибку при выполнении UPDATE запроса
		mock.ExpectExec(`UPDATE categories SET deleted = TRUE WHERE id = \$1`).
			WithArgs(categoryID).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		err := storage.DeleteCategory(ctx, categoryID)

		// Проверка результатов
		require.Error(t, err)
		assert.ErrorContains(t, err, "database error")

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
