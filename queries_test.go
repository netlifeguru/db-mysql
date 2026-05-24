package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/netlifeguru/db"
)

func newQueriesTestConnect() *Connect {
	return &Connect{
		Timeout: time.Second,
		shared: &sharedState{
			Pools:         make(map[string]*Pool),
			sharedPools:   make(map[string]*sharedPool),
			myConnections: make([]string, 0),
		},
	}
}

func TestExecCtxNoConnection(t *testing.T) {
	t.Parallel()

	c := newQueriesTestConnect()

	res, err := c.ExecCtx(context.Background(), db.Query{
		SQL:  "update users set name = ? where id = ?",
		Args: []any{"Martin", 1},
	})

	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result, got %#v", res)
	}
}

func TestExecNoConnection(t *testing.T) {
	t.Parallel()

	c := newQueriesTestConnect()

	res, err := c.Exec(db.Query{
		SQL:  "update users set name = ? where id = ?",
		Args: []any{"Martin", 1},
	})

	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result, got %#v", res)
	}
}

func TestExecCtxInvalidTransactionType(t *testing.T) {
	t.Parallel()

	c := newQueriesTestConnect()
	c.TX = "not-a-sql-tx"

	res, err := c.ExecCtx(context.Background(), db.Query{
		SQL:  "update users set name = ? where id = ?",
		Args: []any{"Martin", 1},
	})

	if !errors.Is(err, db.ErrMissingTx) {
		t.Fatalf("expected ErrMissingTx, got %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result, got %#v", res)
	}
}

func TestExecCtxCancelledContextNoConnection(t *testing.T) {
	t.Parallel()

	c := newQueriesTestConnect()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, err := c.ExecCtx(ctx, db.Query{
		SQL:  "update users set name = ? where id = ?",
		Args: []any{"Martin", 1},
	})

	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection before context error, got %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result, got %#v", res)
	}
}

func TestQueryRowsNoConnection(t *testing.T) {
	t.Parallel()

	c := newQueriesTestConnect()

	rows, err := c.QueryRows(context.Background(), db.Query{SQL: "select * from users"})

	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
	if rows != nil {
		t.Fatalf("expected nil rows, got %#v", rows)
	}
}

func TestQueryPropagatesQueryCtxErrorNoConnection(t *testing.T) {
	t.Parallel()

	c := newQueriesTestConnect()

	err := c.Query(db.Query{SQL: "select * from users"}, func(row map[string]any) error {
		return nil
	})

	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
}

func TestQueryNilCallback(t *testing.T) {
	t.Parallel()

	c := newQueriesTestConnect()

	err := c.Query(db.Query{SQL: "select * from users"}, nil)
	if !errors.Is(err, db.ErrNilEachCallback) {
		t.Fatalf("expected ErrNilEachCallback, got %v", err)
	}
}
