package goninja

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPBearer_Authenticate(t *testing.T) {
	verified := stubUser{id: "bearer-user"}
	verify := func(token string) (User, bool) {
		if token != "good" {
			return nil, false
		}
		return verified, true
	}

	tests := []struct {
		name   string
		verify func(string) (User, bool)
		header string
		want   bool
	}{
		{"valid token calls Verify and authenticates", verify, "Bearer good", true},
		{"missing Authorization header declines", verify, "", false},
		{"wrong scheme declines", verify, "Basic good", false},
		{"empty token declines", verify, "Bearer ", false},
		{"nil Verify declines", nil, "Bearer good", false},
		{"Verify itself declines", func(string) (User, bool) { return nil, false }, "Bearer bad", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := HTTPBearer{Verify: tt.verify}
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				r.Header.Set("Authorization", tt.header)
			}

			user, ok := a.Authenticate(r)
			if ok != tt.want {
				t.Errorf("Authenticate() ok = %v, want %v", ok, tt.want)
			}
			if ok && user.ID() != verified.ID() {
				t.Errorf("Authenticate() user = %v, want %v", user, verified)
			}
		})
	}
}

func TestHTTPBearer_NameAndSecurityScheme(t *testing.T) {
	t.Run("Name defaults to bearer", func(t *testing.T) {
		if got := (HTTPBearer{}).Name(); got != "bearer" {
			t.Errorf("Name() = %q, want %q", got, "bearer")
		}
	})

	t.Run("Name honors AuthName override", func(t *testing.T) {
		if got := (HTTPBearer{AuthName: "custom"}).Name(); got != "custom" {
			t.Errorf("Name() = %q, want %q", got, "custom")
		}
	})

	t.Run("SecurityScheme describes an http bearer scheme", func(t *testing.T) {
		got := (HTTPBearer{}).SecurityScheme()
		if got.Type != "http" || got.Scheme != "bearer" {
			t.Errorf("SecurityScheme() = %+v, want {Type: http, Scheme: bearer}", got)
		}
	})
}

func TestAPIKeyHeader_Authenticate(t *testing.T) {
	verified := stubUser{id: "key-user"}

	verify := func(key string) (User, bool) {
		if key != "secret" {
			return nil, false
		}
		return verified, true
	}

	tests := []struct {
		name       string
		headerName string
		verify     func(string) (User, bool)
		setHeader  string
		setValue   string
		want       bool
	}{
		{"matching header calls Verify and authenticates", "", verify, "X-API-Key", "secret", true},
		{"missing header declines", "", verify, "", "", false},
		{"nil Verify declines", "", nil, "X-API-Key", "secret", false},
		{"Verify itself declines", "", func(string) (User, bool) { return nil, false }, "X-API-Key", "bad", false},
		{"HeaderName override changes which header is read", "X-Custom-Key", verify, "X-Custom-Key", "secret", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := APIKeyHeader{HeaderName: tt.headerName, Verify: tt.verify}
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.setHeader != "" {
				r.Header.Set(tt.setHeader, tt.setValue)
			}

			user, ok := a.Authenticate(r)
			if ok != tt.want {
				t.Errorf("Authenticate() ok = %v, want %v", ok, tt.want)
			}
			if ok && user.ID() != verified.ID() {
				t.Errorf("Authenticate() user = %v, want %v", user, verified)
			}
		})
	}
}

func TestAPIKeyHeader_NameAndSecurityScheme(t *testing.T) {
	t.Run("Name defaults to apiKey", func(t *testing.T) {
		if got := (APIKeyHeader{}).Name(); got != "apiKey" {
			t.Errorf("Name() = %q, want %q", got, "apiKey")
		}
	})

	t.Run("Name honors AuthName override", func(t *testing.T) {
		if got := (APIKeyHeader{AuthName: "custom"}).Name(); got != "custom" {
			t.Errorf("Name() = %q, want %q", got, "custom")
		}
	})

	t.Run("SecurityScheme describes an apiKey header scheme", func(t *testing.T) {
		got := (APIKeyHeader{}).SecurityScheme()
		if got.Type != "apiKey" || got.In != "header" || got.Name != "X-API-Key" {
			t.Errorf("SecurityScheme() = %+v, want {Type: apiKey, In: header, Name: X-API-Key}", got)
		}
	})

	t.Run("SecurityScheme honors HeaderName override", func(t *testing.T) {
		got := (APIKeyHeader{HeaderName: "X-Custom-Key"}).SecurityScheme()
		if got.Name != "X-Custom-Key" {
			t.Errorf("SecurityScheme().Name = %q, want %q", got.Name, "X-Custom-Key")
		}
	})
}

func TestHTTPBasic_Authenticate(t *testing.T) {
	verified := stubUser{id: "basic-user"}
	verify := func(username, password string) (User, bool) {
		if username != "alice" || password != "secret" {
			return nil, false
		}
		return verified, true
	}

	tests := []struct {
		name           string
		verify         func(string, string) (User, bool)
		setCredentials bool
		username       string
		password       string
		want           bool
	}{
		{"valid credentials call Verify and authenticate", verify, true, "alice", "secret", true},
		{"missing Authorization header declines", verify, false, "", "", false},
		{"wrong credentials decline", verify, true, "alice", "wrong", false},
		{"nil Verify declines", nil, true, "alice", "secret", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := HTTPBasic{Verify: tt.verify}
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.setCredentials {
				r.SetBasicAuth(tt.username, tt.password)
			}

			user, ok := a.Authenticate(r)
			if ok != tt.want {
				t.Errorf("Authenticate() ok = %v, want %v", ok, tt.want)
			}
			if ok && user.ID() != verified.ID() {
				t.Errorf("Authenticate() user = %v, want %v", user, verified)
			}
		})
	}
}

func TestHTTPBasic_NameAndSecurityScheme(t *testing.T) {
	t.Run("Name defaults to basic", func(t *testing.T) {
		if got := (HTTPBasic{}).Name(); got != "basic" {
			t.Errorf("Name() = %q, want %q", got, "basic")
		}
	})

	t.Run("Name honors AuthName override", func(t *testing.T) {
		if got := (HTTPBasic{AuthName: "custom"}).Name(); got != "custom" {
			t.Errorf("Name() = %q, want %q", got, "custom")
		}
	})

	t.Run("SecurityScheme describes an http basic scheme", func(t *testing.T) {
		got := (HTTPBasic{}).SecurityScheme()
		if got.Type != "http" || got.Scheme != "basic" {
			t.Errorf("SecurityScheme() = %+v, want {Type: http, Scheme: basic}", got)
		}
	})
}

func TestCookieKey_Authenticate(t *testing.T) {
	verified := stubUser{id: "cookie-user"}
	verify := func(value string) (User, bool) {
		if value != "good" {
			return nil, false
		}
		return verified, true
	}

	tests := []struct {
		name       string
		cookieName string
		verify     func(string) (User, bool)
		setCookie  string
		setValue   string
		want       bool
	}{
		{"matching cookie calls Verify and authenticates", "", verify, "session", "good", true},
		{"missing cookie declines", "", verify, "", "", false},
		{"empty cookie value declines", "", verify, "session", "", false},
		{"nil Verify declines", "", nil, "session", "good", false},
		{"Verify itself declines", "", func(string) (User, bool) { return nil, false }, "session", "bad", false},
		{"CookieName override changes which cookie is read", "custom_session", verify, "custom_session", "good", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := CookieKey{CookieName: tt.cookieName, Verify: tt.verify}
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.setCookie != "" {
				r.AddCookie(&http.Cookie{Name: tt.setCookie, Value: tt.setValue})
			}

			user, ok := a.Authenticate(r)
			if ok != tt.want {
				t.Errorf("Authenticate() ok = %v, want %v", ok, tt.want)
			}
			if ok && user.ID() != verified.ID() {
				t.Errorf("Authenticate() user = %v, want %v", user, verified)
			}
		})
	}
}

func TestCookieKey_NameAndSecurityScheme(t *testing.T) {
	t.Run("Name defaults to cookieAuth", func(t *testing.T) {
		if got := (CookieKey{}).Name(); got != "cookieAuth" {
			t.Errorf("Name() = %q, want %q", got, "cookieAuth")
		}
	})

	t.Run("Name honors AuthName override", func(t *testing.T) {
		if got := (CookieKey{AuthName: "custom"}).Name(); got != "custom" {
			t.Errorf("Name() = %q, want %q", got, "custom")
		}
	})

	t.Run("SecurityScheme describes an apiKey cookie scheme", func(t *testing.T) {
		got := (CookieKey{}).SecurityScheme()
		if got.Type != "apiKey" || got.In != "cookie" || got.Name != "session" {
			t.Errorf("SecurityScheme() = %+v, want {Type: apiKey, In: cookie, Name: session}", got)
		}
	})

	t.Run("SecurityScheme honors CookieName override", func(t *testing.T) {
		got := (CookieKey{CookieName: "custom_session"}).SecurityScheme()
		if got.Name != "custom_session" {
			t.Errorf("SecurityScheme().Name = %q, want %q", got.Name, "custom_session")
		}
	})
}
