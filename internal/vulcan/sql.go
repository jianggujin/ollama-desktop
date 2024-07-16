package vulcan

import (
	"context"
	"database/sql"
)

type SqlDriver struct {
	db *sql.DB
}

func NewSqlDriver(db *sql.DB) *SqlDriver {
	return &SqlDriver{db: db}
}

func (m *SqlDriver) Execute(sql string, values ...interface{}) error {
	_, err := m.db.Exec(sql, values...)
	return err
}

func (m *SqlDriver) ExecuteContext(ctx context.Context, sql string, values ...interface{}) error {
	_, err := m.db.ExecContext(ctx, sql, values...)
	return err
}

func (m *SqlDriver) Query(sql string, values ...interface{}) (*sql.Rows, error) {
	return m.db.Query(sql, values...)
}

func (m *SqlDriver) QueryContext(ctx context.Context, sql string, values ...interface{}) (*sql.Rows, error) {
	return m.db.QueryContext(ctx, sql, values...)
}
