package main

import (
	"strings"
	"testing"
	"time"

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
	if !strings.Contains(sel, "created_on") {
		t.Errorf("list query must use Author.CreatedAt's gorm column, got: %s", sel)
	}
	if !strings.Contains(cnt, "count(*)") {
		t.Errorf("count query was corrupted by the Select: %s", cnt)
	}
}

// TestAuthorListUsesGORMColumnForFilterAndOrder is the runtime counterpart
// to the codegen tests: CreatedAt is publicly named created_at but stored as
// created_on via its gorm:"column:..." tag. Both filtering and ordering must
// use the latter, never the API name.
func TestAuthorListUsesGORMColumnForFilterAndOrder(t *testing.T) {
	db := goninjatest.NewDB(t, &models.Author{}, &models.Book{})
	older := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, author := range []models.Author{
		{ID: "a1", Name: "Older", CreatedAt: older},
		{ID: "a2", Name: "Newer", CreatedAt: newer},
	} {
		if err := db.Create(&author).Error; err != nil {
			t.Fatal(err)
		}
	}

	sqlLog := &sqlCapture{}
	db.Logger = logger.New(sqlLog, logger.Config{LogLevel: logger.Info})
	r := api.NewAuthorResource(db)
	out, total, err := r.List(t.Context(), api.AuthorFilters{
		CreatedAt: &newer,
		Limit:     10,
		Order:     "-created_at",
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(out) != 1 || out[0].Name != "Newer" {
		t.Fatalf("list result = total %d, items %#v; want only Newer", total, out)
	}

	var selectSQL string
	for _, stmt := range sqlLog.stmts {
		if strings.HasPrefix(stmt, "SELECT") && !strings.HasPrefix(stmt, "SELECT count") {
			selectSQL = stmt
		}
	}
	if !strings.Contains(selectSQL, "created_on") || !strings.Contains(selectSQL, "ORDER BY created_on DESC") {
		t.Errorf("filter/order must use gorm column created_on, got: %s", selectSQL)
	}
}
