package mysql

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/netlifeguru/db"
)

func newTxTestConnect() *Connect {
	return &Connect{
		Timeout: time.Second,
		shared: &sharedState{
			Pools:         make(map[string]*Pool),
			sharedPools:   make(map[string]*sharedPool),
			myConnections: make([]string, 0),
		},
	}
}

func TestTransactionCtxNoConnection(t *testing.T) {
	t.Parallel()

	c := newTxTestConnect()

	called := false
	err := c.TransactionCtx(context.Background(), func(tx db.Conn) error {
		called = true
		return nil
	})

	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}

	if !strings.Contains(err.Error(), "transaction") {
		t.Fatalf("expected transaction context in error, got %v", err)
	}

	if called {
		t.Fatalf("transaction callback should not be called")
	}
}

func TestTransactionNoConnection(t *testing.T) {
	t.Parallel()

	c := newTxTestConnect()

	called := false
	err := c.Transaction(func(tx db.Conn) error {
		called = true
		return nil
	})

	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection, got %v", err)
	}

	if !strings.Contains(err.Error(), "transaction") {
		t.Fatalf("expected transaction context in error, got %v", err)
	}

	if called {
		t.Fatalf("transaction callback should not be called")
	}
}

func TestTransactionCtxNoConnectionWithCancelledContext(t *testing.T) {
	t.Parallel()

	c := newTxTestConnect()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	err := c.TransactionCtx(ctx, func(tx db.Conn) error {
		called = true
		return nil
	})

	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection before context error, got %v", err)
	}

	if called {
		t.Fatalf("transaction callback should not be called")
	}
}

func TestTransactionCtxNilCallbackNoConnection(t *testing.T) {
	t.Parallel()

	c := newTxTestConnect()

	err := c.TransactionCtx(context.Background(), nil)

	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection before nil callback panic, got %v", err)
	}
}

func TestTransactionNilCallbackNoConnection(t *testing.T) {
	t.Parallel()

	c := newTxTestConnect()

	err := c.Transaction(nil)

	if !errors.Is(err, db.ErrNoConnection) {
		t.Fatalf("expected ErrNoConnection before nil callback panic, got %v", err)
	}
}
