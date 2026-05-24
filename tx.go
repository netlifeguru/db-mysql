package mysql

import (
	"context"
	"fmt"

	"github.com/netlifeguru/db"
)

func (c *Connect) Transaction(fn func(tx db.Conn) error) (retErr error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout())
	defer cancel()
	return c.TransactionCtx(ctx, fn)
}

func (c *Connect) TransactionCtx(ctx context.Context, fn func(tx db.Conn) error) (retErr error) {
	_, dbConn := c.Connection()
	if dbConn == nil {
		return fmt.Errorf("%w: transaction", db.ErrNoConnection)
	}

	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txConn := &Connect{
		newPoolHook: c.newPoolHook,
		shared:      c.shared,

		Host:       c.Host,
		Identifier: c.Identifier,
		TX:         tx,
		err:        nil,
		Timeout:    c.Timeout,
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			retErr = fmt.Errorf("%w: %v", db.ErrTxPanic, r)
		} else if retErr != nil {
			_ = tx.Rollback()
		} else {
			retErr = tx.Commit()
		}
	}()

	return fn(txConn)
}
