package mysql

import (
	"context"
	"database/sql"

	"github.com/netlifeguru/db"
	"github.com/netlifeguru/mapper"
)

func (c *Connect) QueryCtx(ctx context.Context, query db.Query, each func(row map[string]any) error) error {
	if each == nil {
		return db.ErrNilEachCallback
	}

	sqlRows, err := c.queryRows(ctx, query)
	if err != nil {
		return err
	}
	defer sqlRows.Close()

	return mapper.ScanMapRows(adaptRows(sqlRows), each)
}

func (c *Connect) queryRows(ctx context.Context, query db.Query) (*sql.Rows, error) {
	if c.TX == nil {
		_, dbConn := c.Connection()
		if dbConn == nil {
			return nil, db.ErrNoConnection
		}
		return dbConn.QueryContext(ctx, query.SQL, query.Args...)
	}

	tx, ok := c.TX.(*sql.Tx)
	if !ok {
		return nil, db.ErrMissingTx
	}

	return tx.QueryContext(ctx, query.SQL, query.Args...)
}
