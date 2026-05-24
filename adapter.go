package mysql

import (
	"database/sql"

	"github.com/netlifeguru/mapper"
)

type rowsAdapter struct {
	rows *sql.Rows
}

func adaptRows(rows *sql.Rows) mapper.Rows {
	return rowsAdapter{
		rows: rows,
	}
}

func (r rowsAdapter) Next() bool {
	return r.rows.Next()
}

func (r rowsAdapter) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r rowsAdapter) Err() error {
	return r.rows.Err()
}

func (r rowsAdapter) Close() error {
	return r.rows.Close()
}

func (r rowsAdapter) Columns() ([]string, error) {
	return r.rows.Columns()
}
