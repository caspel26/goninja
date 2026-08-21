// Package goninjatest provides test-only helpers for exercising a
// goninja resource end to end, without hand-rolled httptest/sqlite
// boilerplate in every test file — testing a custom resource should take
// under 10 lines. Kept out of the
// root goninja package (like openapi/docsui/id) since it's optional and
// self-contained: only test code imports it, and it pulls in "testing"
// and an in-memory sqlite driver that a non-test build never needs.
package goninjatest

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// NewDB opens a fresh in-memory SQLite database and AutoMigrates models
// against it, for use as the *gorm.DB a resource under test is
// constructed with. The connection is closed automatically via
// t.Cleanup. Pass every model the resource under test touches, including
// related models reached via Preload — AutoMigrate needs their tables to
// exist too.
func NewDB(t testing.TB, models ...any) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("goninjatest.NewDB: open in-memory sqlite: %v", err)
	}
	if closer, ok := db.ConnPool.(interface{ Close() error }); ok {
		t.Cleanup(func() { _ = closer.Close() })
	}

	if len(models) > 0 {
		if err := db.AutoMigrate(models...); err != nil {
			t.Fatalf("goninjatest.NewDB: automigrate: %v", err)
		}
	}

	return db
}
