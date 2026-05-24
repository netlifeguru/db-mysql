package mysql

import (
	"fmt"

	"github.com/netlifeguru/db"
)

func (c *Connect) AnalyzeSQL(sql string) (prepared string, placeholders int, err error) {
	var (
		inSingle       bool
		inDouble       bool
		inBacktick     bool
		inLineComment  bool
		inBlockComment bool
	)

	i := 0
	for i < len(sql) {
		ch := sql[i]

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
			}
			i++
			continue
		}

		if inBlockComment {
			if ch == '*' && i+1 < len(sql) && sql[i+1] == '/' {
				i += 2
				inBlockComment = false
				continue
			}
			i++
			continue
		}

		if !inSingle && !inDouble && !inBacktick {
			if ch == '-' && i+1 < len(sql) && sql[i+1] == '-' {
				if i+2 == len(sql) || isMySQLSpace(sql[i+2]) {
					inLineComment = true
					i += 2
					continue
				}
			}

			if ch == '#' {
				inLineComment = true
				i++
				continue
			}

			if ch == '/' && i+1 < len(sql) && sql[i+1] == '*' {
				inBlockComment = true
				i += 2
				continue
			}
		}

		if ch == '\'' && !inDouble && !inBacktick {
			if inSingle {
				if i+1 < len(sql) && sql[i+1] == '\'' {
					i += 2
					continue
				}

				if i > 0 && sql[i-1] == '\\' {
					i++
					continue
				}

				inSingle = false
			} else {
				inSingle = true
			}

			i++
			continue
		}

		if ch == '"' && !inSingle && !inBacktick {
			if inDouble {
				if i+1 < len(sql) && sql[i+1] == '"' {
					i += 2
					continue
				}

				if i > 0 && sql[i-1] == '\\' {
					i++
					continue
				}

				inDouble = false
			} else {
				inDouble = true
			}

			i++
			continue
		}

		if ch == '`' && !inSingle && !inDouble {
			if inBacktick {
				if i+1 < len(sql) && sql[i+1] == '`' {
					i += 2
					continue
				}

				inBacktick = false
			} else {
				inBacktick = true
			}

			i++
			continue
		}

		if ch == '?' && !inSingle && !inDouble && !inBacktick {
			placeholders++
		}

		i++
	}

	switch {
	case inSingle:
		return "", 0, fmt.Errorf("%w near position %d", db.ErrPrepareUnterminatedSingleQuotedString, i)
	case inDouble:
		return "", 0, fmt.Errorf("%w near position %d", db.ErrPrepareUnterminatedDoubleQuotedIdent, i)
	case inBacktick:
		return "", 0, fmt.Errorf("unterminated backtick identifier near position %d", i)
	case inBlockComment:
		return "", 0, fmt.Errorf("%w near position %d", db.ErrPrepareUnterminatedBlockComment, i)
	}

	return sql, placeholders, nil
}

func isMySQLSpace(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	default:
		return false
	}
}

func (c *Connect) SelectSQL(q db.DialectSQL) string {
	return q.Mysql
}
