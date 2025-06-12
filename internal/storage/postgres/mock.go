package postgres

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MockPgxPool представляет собой упрощенную имитацию пула соединений pgx для тестирования
type MockPgxPool struct {
	db *sql.DB
}

// DB возвращает внутренний *sql.DB для доступа к базе данных
func (p *MockPgxPool) DB() *sql.DB {
	return p.db
}

// Exec выполняет SQL-запрос без возврата строк
func (p *MockPgxPool) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	_, err := p.db.ExecContext(ctx, sql, arguments...)
	return pgconn.CommandTag{}, err
}

// Query выполняет запрос, который возвращает строки
func (p *MockPgxPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	// В контексте тестирования нам не нужна полная реализация Rows
	// Go-sqlmock будет проверять вызовы, но не результаты
	return nil, nil
}

// QueryRow выполняет запрос, который возвращает максимум одну строку
func (p *MockPgxPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	// В контексте тестирования нам не нужна полная реализация Row
	// Go-sqlmock будет проверять вызовы, но не результаты
	return nil
}

// Close закрывает пул соединений
func (p *MockPgxPool) Close() {
	p.db.Close()
}

// NewMockPgxPool создает новый мок pgxpool.Pool для тестирования
func NewMockPgxPool(db *sql.DB) *MockPgxPool {
	return &MockPgxPool{
		db: db,
	}
}
