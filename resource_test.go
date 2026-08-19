package goninja

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type resourceTestRow struct {
	ID   uint
	Name string
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	if err := db.AutoMigrate(&resourceTestRow{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func TestWithTx_TxFromContext(t *testing.T) {
	db := newTestDB(t)

	if _, ok := TxFromContext(context.Background()); ok {
		t.Error("TxFromContext on a bare context: ok = true, want false")
	}

	ctx := WithTx(context.Background(), db)
	tx, ok := TxFromContext(ctx)
	if !ok {
		t.Fatal("TxFromContext: ok = false, want true")
	}
	if tx != db {
		t.Error("TxFromContext did not return the same *gorm.DB passed to WithTx")
	}
}

func TestBaseResource_DB_FallsBackToBaseConnection(t *testing.T) {
	db := newTestDB(t)
	var r BaseResource
	r.SetDB(db)

	got := r.DB(context.Background())
	if got == nil {
		t.Fatal("DB(ctx) returned nil")
	}
	if err := got.AutoMigrate(&resourceTestRow{}); err != nil {
		t.Errorf("DB(ctx) is not a usable *gorm.DB: %v", err)
	}
}

func TestBaseResource_DB_UsesContextTransaction(t *testing.T) {
	db := newTestDB(t)
	var r BaseResource
	r.SetDB(db)

	var txSeen *gorm.DB
	err := db.Transaction(func(tx *gorm.DB) error {
		ctx := WithTx(context.Background(), tx)
		txSeen = r.DB(ctx)
		return nil
	})
	if err != nil {
		t.Fatalf("Transaction: %v", err)
	}
	if txSeen == nil {
		t.Fatal("DB(ctx) inside a transaction returned nil")
	}
}

func TestInTransaction_CommitsOnSuccess(t *testing.T) {
	db := newTestDB(t)

	_, err := InTransaction(context.Background(), db, func(ctx context.Context) (int, error) {
		tx, _ := TxFromContext(ctx)
		return 0, tx.Create(&resourceTestRow{Name: "committed"}).Error
	})
	if err != nil {
		t.Fatalf("InTransaction: %v", err)
	}

	var count int64
	db.Model(&resourceTestRow{}).Count(&count)
	if count != 1 {
		t.Errorf("row count after successful InTransaction = %d, want 1", count)
	}
}

func TestInTransaction_RollsBackOnError(t *testing.T) {
	db := newTestDB(t)
	wantErr := errors.New("boom")

	_, err := InTransaction(context.Background(), db, func(ctx context.Context) (int, error) {
		tx, _ := TxFromContext(ctx)
		if err := tx.Create(&resourceTestRow{Name: "rolled back"}).Error; err != nil {
			return 0, err
		}
		return 0, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("InTransaction err = %v, want %v", err, wantErr)
	}

	var count int64
	db.Model(&resourceTestRow{}).Count(&count)
	if count != 0 {
		t.Errorf("row count after failed InTransaction = %d, want 0 (rolled back)", count)
	}
}

func TestBaseResource_SetSelf_Self(t *testing.T) {
	var r BaseResource
	if r.Self() != nil {
		t.Error("Self() before SetSelf = non-nil, want nil")
	}

	type wrapper struct{}
	w := &wrapper{}
	r.SetSelf(w)
	if r.Self() != any(w) {
		t.Error("Self() after SetSelf did not return the value passed in")
	}
}

func TestBaseResource_Protect(t *testing.T) {
	var authRan, globalRan bool
	authMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authRan = true
			next.ServeHTTP(w, r)
		})
	}
	globalMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			globalRan = true
			next.ServeHTTP(w, r)
		})
	}

	var r BaseResource
	r.SetConfig(Config{
		DefaultAuth: AuthPolicy{
			Protected:  []string{"create"},
			Middleware: []func(http.Handler) http.Handler{authMW},
		},
		Middleware: []func(http.Handler) http.Handler{globalMW},
	})

	handler := func(w http.ResponseWriter, req *http.Request) { w.WriteHeader(http.StatusOK) }

	t.Run("protected route runs auth and global middleware", func(t *testing.T) {
		authRan, globalRan = false, false
		h := r.Protect("create", ResourceConfig{}, handler)
		h(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))
		if !authRan || !globalRan {
			t.Errorf("authRan=%v globalRan=%v, want both true", authRan, globalRan)
		}
	})

	t.Run("unprotected route only runs global middleware", func(t *testing.T) {
		authRan, globalRan = false, false
		h := r.Protect("list", ResourceConfig{}, handler)
		h(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		if authRan {
			t.Error("authRan = true for an unprotected route, want false")
		}
		if !globalRan {
			t.Error("globalRan = false, want true (global middleware always runs)")
		}
	})

	t.Run("AlsoProtect additively protects a route", func(t *testing.T) {
		authRan, globalRan = false, false
		h := r.Protect("retrieve", ResourceConfig{Auth: AuthOverride{AlsoProtect: []string{"retrieve"}}}, handler)
		h(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		if !authRan {
			t.Error("authRan = false for a route added via AlsoProtect, want true")
		}
	})

	t.Run("Public punches a hole in the default protection", func(t *testing.T) {
		authRan, globalRan = false, false
		h := r.Protect("create", ResourceConfig{Auth: AuthOverride{Public: []string{"create"}}}, handler)
		h(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))
		if authRan {
			t.Error("authRan = true for a route in Auth.Public, want false")
		}
		if !globalRan {
			t.Error("globalRan = false, want true (global middleware still runs on a public route)")
		}
	})
}

func TestBaseResource_ErrorMapper_DefaultsWhenUnset(t *testing.T) {
	var r BaseResource
	if _, ok := r.ErrorMapper().(DefaultErrorMapper); !ok {
		t.Errorf("ErrorMapper() before SetErrorMapper = %T, want DefaultErrorMapper", r.ErrorMapper())
	}

	m := DefaultErrorMapper{}
	r.SetErrorMapper(m)
	if r.ErrorMapper() != ErrorMapper(m) {
		t.Error("ErrorMapper() after SetErrorMapper did not return the value passed in")
	}
}

func TestBaseResource_OpenAPITags(t *testing.T) {
	var r BaseResource
	if r.OpenAPITags() != nil {
		t.Errorf("OpenAPITags() before SetOpenAPITags = %v, want nil", r.OpenAPITags())
	}

	r.SetOpenAPITags("Books", "Catalog")
	got := r.OpenAPITags()
	if len(got) != 2 || got[0] != "Books" || got[1] != "Catalog" {
		t.Errorf("OpenAPITags() = %v, want [Books Catalog]", got)
	}
}

func TestBaseResource_ExcludeFromDocs_DocsExcluded(t *testing.T) {
	var r BaseResource
	if r.DocsExcluded() {
		t.Error("DocsExcluded() before ExcludeFromDocs = true, want false")
	}
	r.ExcludeFromDocs()
	if !r.DocsExcluded() {
		t.Error("DocsExcluded() after ExcludeFromDocs = false, want true")
	}
}

func TestBaseResource_Config(t *testing.T) {
	var r BaseResource
	if r.Config().DefaultAuth.Protected != nil {
		t.Error("Config() before SetConfig is not the zero value")
	}

	cfg := Config{DefaultAuth: AuthPolicy{Protected: []string{"create"}}}
	r.SetConfig(cfg)
	if len(r.Config().DefaultAuth.Protected) != 1 || r.Config().DefaultAuth.Protected[0] != "create" {
		t.Errorf("Config() after SetConfig = %+v, want %+v", r.Config(), cfg)
	}
}

func TestBaseResource_Actions(t *testing.T) {
	var r BaseResource
	if got := r.Actions(); got != nil {
		t.Errorf("Actions() before SetActions = %v, want nil", got)
	}

	a := Action{Name: "publish", Method: "POST", UrlPath: "publish"}
	r.SetActions(a)
	got := r.Actions()
	if len(got) != 1 || got[0].Name != "publish" {
		t.Errorf("Actions() after SetActions = %+v, want [%+v]", got, a)
	}
}

func TestBaseResource_Protect_ZeroConfigIsNoOp(t *testing.T) {
	var r BaseResource // never SetConfig — plain Mount path
	var ran bool
	handler := func(w http.ResponseWriter, req *http.Request) { ran = true }

	h := r.Protect("create", ResourceConfig{}, handler)
	h(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))

	if !ran {
		t.Error("handler did not run through a zero-value Protect")
	}
}
