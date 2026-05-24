package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/netlifeguru/db"
)

func newScanTestConnect() *Connect {
	return &Connect{
		Timeout: time.Second,
		shared: &sharedState{
			Pools:         make(map[string]*Pool),
			sharedPools:   make(map[string]*sharedPool),
			myConnections: make([]string, 0),
		},
	}
}

func TestQueryCtxNilCallback(t *testing.T) {
	t.Parallel()

	c := newScanTestConnect()

	err := c.QueryCtx(context.Background(), db.Query{SQL: "select * from users"}, nil)
	if !errors.Is(err, db.ErrNilEachCallback) {
		t.Fatalf("expected ErrNilEachCallback, got %v", err)
	}
}

func TestQueryCtxNoConnection(t *testing.T) {
	t.Parallel()

	c := newScanTestConnect()

	err := c.QueryCtx(context.Background(), db.Query{SQL: "select * from users"}, func(row map[string]any) error {
		return nil
	})

	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
}

func TestQueryCtxInvalidTransactionType(t *testing.T) {
	t.Parallel()

	c := newScanTestConnect()
	c.TX = "not-a-sql-tx"

	err := c.QueryCtx(context.Background(), db.Query{SQL: "select * from users"}, func(row map[string]any) error {
		return nil
	})

	if !errors.Is(err, db.ErrMissingTx) {
		t.Fatalf("expected ErrMissingTx, got %v", err)
	}
}

func TestScanQueryRowsNoConnection(t *testing.T) {
	t.Parallel()

	c := newScanTestConnect()

	rows, err := c.queryRows(context.Background(), db.Query{SQL: "select * from users"})
	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}
	if rows != nil {
		t.Fatalf("expected nil rows, got %#v", rows)
	}
}

func TestQueryRowsInvalidTransactionType(t *testing.T) {
	t.Parallel()

	c := newScanTestConnect()
	c.TX = "not-a-sql-tx"

	rows, err := c.queryRows(context.Background(), db.Query{SQL: "select * from users"})
	if !errors.Is(err, db.ErrMissingTx) {
		t.Fatalf("expected ErrMissingTx, got %v", err)
	}
	if rows != nil {
		t.Fatalf("expected nil rows, got %#v", rows)
	}
}

func TestQueryCtxNilCallbackTakesPrecedenceOverNoConnection(t *testing.T) {
	t.Parallel()

	c := newScanTestConnect()

	err := c.QueryCtx(context.Background(), db.Query{SQL: "select * from users"}, nil)
	if !errors.Is(err, db.ErrNilEachCallback) {
		t.Fatalf("expected ErrNilEachCallback, got %v", err)
	}
}

func TestQueryCtxNilCallbackTakesPrecedenceOverInvalidTransaction(t *testing.T) {
	t.Parallel()

	c := newScanTestConnect()
	c.TX = "not-a-sql-tx"

	err := c.QueryCtx(context.Background(), db.Query{SQL: "select * from users"}, nil)
	if !errors.Is(err, db.ErrNilEachCallback) {
		t.Fatalf("expected ErrNilEachCallback, got %v", err)
	}
}
