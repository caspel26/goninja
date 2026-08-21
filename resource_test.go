package goninja

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/caspel26/goninja/openapi"
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

type fakeAuthenticator struct {
	name    string
	allow   bool
	authRan *bool
}

func (a *fakeAuthenticator) Authenticate(r *http.Request) (User, bool) {
	if a.authRan != nil {
		*a.authRan = true
	}
	if !a.allow {
		return nil, false
	}
	return stubUser{id: a.name}, true
}
func (a *fakeAuthenticator) Name() string { return a.name }
func (a *fakeAuthenticator) SecurityScheme() openapi.SecurityScheme {
	return openapi.SecurityScheme{Type: "http", Scheme: "bearer"}
}

type stubUser struct{ id string }

func (u stubUser) ID() string { return u.id }

func TestBaseResource_Protect(t *testing.T) {
	var authRan, globalRan bool
	auth := &fakeAuthenticator{name: "test", allow: true, authRan: &authRan}
	globalMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			globalRan = true
			next.ServeHTTP(w, r)
		})
	}

	var r BaseResource
	r.SetConfig(Config{
		DefaultAuth: AuthPolicy{
			Routes: []Route{RouteCreate},
			Auth:   []Authenticator{auth},
		},
		Middleware: []func(http.Handler) http.Handler{globalMW},
	})

	handler := func(w http.ResponseWriter, req *http.Request) { w.WriteHeader(http.StatusOK) }

	t.Run("protected route runs auth and global middleware", func(t *testing.T) {
		authRan, globalRan = false, false
		h := r.Protect(RouteCreate, ResourceConfig{}, handler)
		h(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))
		if !authRan || !globalRan {
			t.Errorf("authRan=%v globalRan=%v, want both true", authRan, globalRan)
		}
	})

	t.Run("unprotected route only runs global middleware", func(t *testing.T) {
		authRan, globalRan = false, false
		h := r.Protect(RouteList, ResourceConfig{}, handler)
		h(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		if authRan {
			t.Error("authRan = true for an unprotected route, want false")
		}
		if !globalRan {
			t.Error("globalRan = false, want true (global middleware always runs)")
		}
	})

}

func TestBaseResource_Protect_ResourceConfigOverrides(t *testing.T) {
	var authRan, globalRan bool
	auth := &fakeAuthenticator{name: "test", allow: true, authRan: &authRan}
	globalMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			globalRan = true
			next.ServeHTTP(w, r)
		})
	}

	var r BaseResource
	r.SetConfig(Config{
		DefaultAuth: AuthPolicy{
			Routes: []Route{RouteCreate},
			Auth:   []Authenticator{auth},
		},
		Middleware: []func(http.Handler) http.Handler{globalMW},
	})

	handler := func(w http.ResponseWriter, req *http.Request) { w.WriteHeader(http.StatusOK) }

	t.Run("per-route override additively protects a route", func(t *testing.T) {
		authRan, globalRan = false, false
		h := r.Protect(RouteRetrieve, ResourceConfig{Auth: map[Route]RouteAuth{RouteRetrieve: {}}}, handler)
		h(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		if !authRan {
			t.Error("authRan = false for a route added via a RouteAuth override, want true")
		}
	})

	t.Run("Public punches a hole in the default protection", func(t *testing.T) {
		authRan, globalRan = false, false
		h := r.Protect(RouteCreate, ResourceConfig{Auth: map[Route]RouteAuth{RouteCreate: {Public: true}}}, handler)
		h(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))
		if authRan {
			t.Error("authRan = true for a route marked Public, want false")
		}
		if !globalRan {
			t.Error("globalRan = false, want true (global middleware still runs on a public route)")
		}
	})

	t.Run("override with its own Auth replaces the default authenticators", func(t *testing.T) {
		authRan, globalRan = false, false
		var ownAuthRan bool
		own := &fakeAuthenticator{name: "own", allow: true, authRan: &ownAuthRan}
		h := r.Protect(RouteCreate, ResourceConfig{Auth: map[Route]RouteAuth{RouteCreate: {Auth: []Authenticator{own}}}}, handler)
		h(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))
		if authRan {
			t.Error("authRan = true, want the override's own Authenticator used instead of the default")
		}
		if !ownAuthRan {
			t.Error("the override's own Authenticator did not run")
		}
	})
}

func TestBaseResource_SecurityFor(t *testing.T) {
	auth := &fakeAuthenticator{name: "test", allow: true}
	var r BaseResource
	r.SetConfig(Config{DefaultAuth: AuthPolicy{Routes: []Route{RouteCreate}, Auth: []Authenticator{auth}}})

	t.Run("protected route returns a requirement and its scheme", func(t *testing.T) {
		reqs, schemes := r.SecurityFor(RouteCreate, ResourceConfig{})
		if len(reqs) != 1 || len(reqs[0]["test"]) != 0 {
			t.Errorf("reqs = %+v, want one entry for %q", reqs, "test")
		}
		if _, ok := schemes["test"]; !ok {
			t.Errorf("schemes = %+v, want an entry for %q", schemes, "test")
		}
	})

	t.Run("unprotected route returns nil", func(t *testing.T) {
		reqs, schemes := r.SecurityFor(RouteList, ResourceConfig{})
		if reqs != nil || schemes != nil {
			t.Errorf("SecurityFor on an unprotected route = %v, %v, want nil, nil", reqs, schemes)
		}
	})

	t.Run("protected route with no resolved Authenticators returns nil", func(t *testing.T) {
		var empty BaseResource
		empty.SetConfig(Config{DefaultAuth: AuthPolicy{Routes: []Route{RouteCreate}}})
		reqs, schemes := empty.SecurityFor(RouteCreate, ResourceConfig{})
		if reqs != nil || schemes != nil {
			t.Errorf("SecurityFor with no Authenticators = %v, %v, want nil, nil", reqs, schemes)
		}
	})
}

func TestRequireAuth_TriesEachAuthenticatorInOrder(t *testing.T) {
	var ran bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { ran = true })

	t.Run("falls through to a later Authenticator that accepts", func(t *testing.T) {
		ran = false
		declining := &fakeAuthenticator{name: "declining", allow: false}
		accepting := &fakeAuthenticator{name: "accepting", allow: true}
		h := requireAuth([]Authenticator{declining, accepting}, handler)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		if !ran {
			t.Error("handler did not run even though the second Authenticator accepted")
		}
	})

	t.Run("responds 401 once every Authenticator declines", func(t *testing.T) {
		ran = false
		declining := &fakeAuthenticator{name: "declining", allow: false}
		h := requireAuth([]Authenticator{declining}, handler)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		if ran {
			t.Error("handler ran despite every Authenticator declining")
		}
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
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

func TestBaseResource_ErrorMapper_FallsBackToConfigDefault(t *testing.T) {
	var r BaseResource
	global := NewErrorMapper(NewErrorMapping(func(err error) (int, any) {
		return http.StatusTeapot, nil
	}))
	r.SetConfig(Config{DefaultErrorMapper: global})

	status, _ := r.ErrorMapper().MapError(errors.New("boom"))
	if status != http.StatusTeapot {
		t.Errorf("status = %d, want %d (Config.DefaultErrorMapper)", status, http.StatusTeapot)
	}
}

func TestBaseResource_ErrorMapper_OwnMapperBeatsConfigDefault(t *testing.T) {
	var r BaseResource
	r.SetConfig(Config{DefaultErrorMapper: NewErrorMapper(NewErrorMapping(func(err error) (int, any) {
		return http.StatusTeapot, nil
	}))})
	r.SetErrorMapper(DefaultErrorMapper{})

	status, _ := r.ErrorMapper().MapError(errors.New("boom"))
	if status != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d (the resource's own SetErrorMapper value)", status, http.StatusInternalServerError)
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
	if r.Config().DefaultAuth.Routes != nil {
		t.Error("Config() before SetConfig is not the zero value")
	}

	cfg := Config{DefaultAuth: AuthPolicy{Routes: []Route{RouteCreate}}}
	r.SetConfig(cfg)
	if len(r.Config().DefaultAuth.Routes) != 1 || r.Config().DefaultAuth.Routes[0] != RouteCreate {
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

	h := r.Protect(RouteCreate, ResourceConfig{}, handler)
	h(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))

	if !ran {
		t.Error("handler did not run through a zero-value Protect")
	}
}
