package mysql

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/netlifeguru/mapper"
)

func init() {
	sql.Register("mysql_adapter_test_driver", fakeDriver{})
}

func TestAdaptRowsImplementsMapperRows(t *testing.T) {
	dbConn, err := sql.Open("mysql_adapter_test_driver", "")
	if err != nil {
		t.Fatalf("sql.Open error: %v", err)
	}
	defer dbConn.Close()

	rows, err := dbConn.Query("SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("db.Query error: %v", err)
	}
	defer rows.Close()

	adapted := adaptRows(rows)

	var _ mapper.Rows = adapted
}

func TestRowsAdapterColumns(t *testing.T) {
	dbConn, err := sql.Open("mysql_adapter_test_driver", "")
	if err != nil {
		t.Fatalf("sql.Open error: %v", err)
	}
	defer dbConn.Close()

	rows, err := dbConn.Query("SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("db.Query error: %v", err)
	}
	defer rows.Close()

	adapted := adaptRows(rows)

	got, err := adapted.Columns()
	if err != nil {
		t.Fatalf("Columns error: %v", err)
	}

	want := []string{"id", "name"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected columns %#v, got %#v", want, got)
	}
}

func TestRowsAdapterNextScanErr(t *testing.T) {
	dbConn, err := sql.Open("mysql_adapter_test_driver", "")
	if err != nil {
		t.Fatalf("sql.Open error: %v", err)
	}
	defer dbConn.Close()

	rows, err := dbConn.Query("SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("db.Query error: %v", err)
	}
	defer rows.Close()

	adapted := adaptRows(rows)

	var results []string
	for adapted.Next() {
		var id string
		var name string

		if err := adapted.Scan(&id, &name); err != nil {
			t.Fatalf("Scan error: %v", err)
		}

		results = append(results, id+":"+name)
	}

	if err := adapted.Err(); err != nil {
		t.Fatalf("Err returned error: %v", err)
	}

	want := []string{"1:John", "2:Jane"}
	if !reflect.DeepEqual(results, want) {
		t.Fatalf("expected results %#v, got %#v", want, results)
	}
}

func TestRowsAdapterClose(t *testing.T) {
	dbConn, err := sql.Open("mysql_adapter_test_driver", "")
	if err != nil {
		t.Fatalf("sql.Open error: %v", err)
	}
	defer dbConn.Close()

	rows, err := dbConn.Query("SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("db.Query error: %v", err)
	}

	adapted := adaptRows(rows)
	if err := adapted.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

type fakeDriver struct{}

func (fakeDriver) Open(name string) (driver.Conn, error) {
	return fakeConn{}, nil
}

type fakeConn struct{}

func (fakeConn) Prepare(query string) (driver.Stmt, error) {
	return fakeStmt{}, nil
}

func (fakeConn) Close() error { return nil }

func (fakeConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not supported in fake driver")
}

type fakeStmt struct{}

func (fakeStmt) Close() error  { return nil }
func (fakeStmt) NumInput() int { return -1 }

func (fakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, errors.New("exec is not supported in fake driver")
}

func (fakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &fakeRows{
		columns: []string{"id", "name"},
		values: [][]driver.Value{
			{"1", "John"},
			{"2", "Jane"},
		},
	}, nil
}

type fakeRows struct {
	columns []string
	values  [][]driver.Value
	index   int
	closed  bool
}

func (r *fakeRows) Columns() []string {
	out := make([]string, len(r.columns))
	copy(out, r.columns)
	return out
}

func (r *fakeRows) Close() error {
	r.closed = true
	return nil
}

func (r *fakeRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}

	row := r.values[r.index]
	r.index++
	copy(dest, row)
	return nil
}
