package main

import (
	"strings"
	"testing"

	"github.com/caspel26/goninja/examples/prototype/internal/api"
	"github.com/caspel26/goninja/examples/prototype/models"
	"github.com/caspel26/goninja/goninjatest"
	"gorm.io/gorm/logger"
)

// sqlCapture is a minimal gorm/logger.Writer that records every SQL
// statement gorm traces, so a test can assert on the exact query a
// generated List actually issues.
type sqlCapture struct{ stmts []string }

func (c *sqlCapture) Printf(_ string, args ...any) {
	// gorm's Info-level Trace calls Printf(traceStr, fileWithLineNum,
	// elapsedMs, rows, sql) — sql is always the last arg.
	if len(args) == 0 {
		return
	}
	if s, ok := args[len(args)-1].(string); ok {
		c.stmts = append(c.stmts, s)
	}
}

// TestListSelectsOnlyListColumns proves the ListSelectColumns/Select
// optimization (ir.go, model.go.tmpl): a generated List issues a SELECT
// naming only the <Model>List fields, never SELECT *, and Count still
// works correctly alongside it. Author.Bio is retrieve-only (not on
// AuthorList), so it's the field this test uses to prove the column got
// dropped from the query, not just from the response.
func TestListSelectsOnlyListColumns(t *testing.T) {
	db := goninjatest.NewDB(t, &models.Author{}, &models.Book{})
	db.Create(&models.Author{ID: "a1", Name: "N", Bio: "a very long bio"})

	sqlLog := &sqlCapture{}
	db.Logger = logger.New(sqlLog, logger.Config{LogLevel: logger.Info})

	r := api.NewAuthorResource(db)
	out, total, err := r.List(t.Context(), api.AuthorFilters{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(out) != 1 {
		t.Fatalf("total=%d len=%d, want 1/1 — Count broken by Select?", total, len(out))
	}

	var sel, cnt string
	for _, s := range sqlLog.stmts {
		switch {
		case strings.HasPrefix(s, "SELECT count"):
			cnt = s
		case strings.HasPrefix(s, "SELECT"):
			sel = s
		}
	}

	if strings.Contains(sel, "bio") || strings.Contains(sel, "`*`") || strings.Contains(sel, "SELECT * ") {
		t.Errorf("list query still reads columns the list schema drops: %s", sel)
	}
	if !strings.Contains(cnt, "count(*)") {
		t.Errorf("count query was corrupted by the Select: %s", cnt)
	}
}
