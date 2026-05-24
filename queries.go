package mysql

import (
	"context"
	"database/sql"

	"github.com/netlifeguru/db"
	"github.com/netlifeguru/mapper"
)

func (c *Connect) Exec(q db.Query) (db.Result, error) {
	return c.ExecCtx(context.Background(), q)
}

func (c *Connect) ExecCtx(ctx context.Context, q db.Query) (db.Result, error) {
	if c.TX != nil {
		tx, ok := c.TX.(*sql.Tx)
		if !ok {
			return nil, db.ErrMissingTx
		}

		res, err := tx.ExecContext(ctx, q.SQL, q.Args...)
		if err != nil {
			return nil, err
		}

		lastInsertId, err := res.LastInsertId()

		if err != nil {
			return nil, err
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return nil, err
		}

		return Result{
			lastInsertId: lastInsertId,
			rowsAffected: rowsAffected,
		}, nil
	}

	_, dbConn := c.Connection()
	if dbConn == nil {
		return nil, db.ErrNoConnection
	}

	res, err := dbConn.ExecContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}

	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	return Result{
		lastInsertId: lastInsertId,
		rowsAffected: rowsAffected,
	}, nil
}

func (c *Connect) Query(query db.Query, each func(row map[string]any) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout())
	defer cancel()

	return c.QueryCtx(ctx, query, each)
}

func (c *Connect) QueryRows(ctx context.Context, q db.Query) (mapper.Rows, error) {
	rows, err := c.queryRows(ctx, q)
	if err != nil {
		return nil, err
	}

	return adaptRows(rows), nil
}
