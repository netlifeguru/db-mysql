package mysql

import (
	"errors"
	"strings"
	"testing"

	"github.com/netlifeguru/db"
)

func TestAnalyzeSQLCountsPlaceholders(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	prepared, placeholders, err := c.AnalyzeSQL(
		"select * from users where id = ? and email = ? and active = ?",
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if prepared != "select * from users where id = ? and email = ? and active = ?" {
		t.Fatalf("unexpected prepared query: %q", prepared)
	}

	if placeholders != 3 {
		t.Fatalf("expected 3 placeholders, got %d", placeholders)
	}
}

func TestAnalyzeSQLNoPlaceholders(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	prepared, placeholders, err := c.AnalyzeSQL("select * from users")
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if prepared != "select * from users" {
		t.Fatalf("unexpected prepared query: %q", prepared)
	}

	if placeholders != 0 {
		t.Fatalf("expected 0 placeholders, got %d", placeholders)
	}
}

func TestAnalyzeSQLIgnoresQuestionMarkInSingleQuotedString(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		"select * from users where text = 'hello ? world' and id = ?",
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLIgnoresQuestionMarkInDoubleQuotedString(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		`select * from users where text = "hello ? world" and id = ?`,
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLIgnoresQuestionMarkInBacktickIdentifier(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		"select `weird ? column` from users where id = ?",
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLIgnoresQuestionMarkInLineComment(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		"select * from users -- ignored ?\nwhere id = ?",
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLIgnoresQuestionMarkInHashComment(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		"select * from users # ignored ?\nwhere id = ?",
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLIgnoresQuestionMarkInBlockComment(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		"select * from users /* ignored ? */ where id = ?",
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLQuestionMarksInMultipleContexts(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(`
select ` + "`ignored ?`" + `, '?', "?"
from users
where id = ?
  and name = ?
-- ignored ?
# ignored ?
/* ignored ? */
`)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 2 {
		t.Fatalf("expected 2 placeholders, got %d", placeholders)
	}
}

func TestAnalyzeSQLEscapedSingleQuote(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		`select * from users where text = 'john\'s ? text' and id = ?`,
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLDoubledSingleQuote(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		`select * from users where text = 'john''s ? text' and id = ?`,
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLEscapedDoubleQuote(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		`select * from users where text = "hello \" ? world" and id = ?`,
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLDoubledDoubleQuote(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		`select * from users where text = "hello "" ? world" and id = ?`,
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLDoubledBacktick(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL(
		"select `weird `` ? column` from users where id = ?",
	)
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestAnalyzeSQLUnterminatedSingleQuotedString(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, _, err := c.AnalyzeSQL("select * from users where name = 'john")
	if !errors.Is(err, db.ErrPrepareUnterminatedSingleQuotedString) {
		t.Fatalf("expected ErrPrepareUnterminatedSingleQuotedString, got %v", err)
	}
}

func TestAnalyzeSQLUnterminatedDoubleQuotedIdent(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, _, err := c.AnalyzeSQL(`select * from users where name = "john`)
	if !errors.Is(err, db.ErrPrepareUnterminatedDoubleQuotedIdent) {
		t.Fatalf("expected ErrPrepareUnterminatedDoubleQuotedIdent, got %v", err)
	}
}

func TestAnalyzeSQLUnterminatedBacktickIdentifier(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, _, err := c.AnalyzeSQL("select `name from users")
	if err == nil {
		t.Fatalf("expected error")
	}

	if !strings.Contains(err.Error(), "unterminated backtick identifier") {
		t.Fatalf("expected unterminated backtick identifier error, got %v", err)
	}
}

func TestAnalyzeSQLUnterminatedBlockComment(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, _, err := c.AnalyzeSQL("select * from users /* comment ?")
	if !errors.Is(err, db.ErrPrepareUnterminatedBlockComment) {
		t.Fatalf("expected ErrPrepareUnterminatedBlockComment, got %v", err)
	}
}

func TestAnalyzeSQLLineCommentAtEOF(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	_, placeholders, err := c.AnalyzeSQL("select * from users where id = ? -- ignored ?")
	if err != nil {
		t.Fatalf("AnalyzeSQL returned error: %v", err)
	}

	if placeholders != 1 {
		t.Fatalf("expected 1 placeholder, got %d", placeholders)
	}
}

func TestSelectSQL(t *testing.T) {
	t.Parallel()

	c := &Connect{}

	got := c.SelectSQL(db.DialectSQL{
		Postgresql: "select $1",
		Mysql:      "select ?",
		Scylla:     "select scylla",
	})

	if got != "select ?" {
		t.Fatalf("expected mysql SQL, got %q", got)
	}
}
