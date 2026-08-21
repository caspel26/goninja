// Ready-made Authenticator implementations for common auth schemes,
// mirroring Django Ninja's
// HttpBearer/ApiKeyHeader base classes: the request-parsing/OpenAPI-scheme
// boilerplate is handled here, and the caller only supplies the Verify
// closure that turns a raw token/key into a User. A caller with more
// unusual requirements (a custom header, a different scheme, combining
// multiple credentials) still implements Authenticator directly — these
// are conveniences, not the only way in.
package goninja

import (
	"net/http"
	"strings"

	"github.com/caspel26/goninja/openapi"
)

// HTTPBearer is a ready-made Authenticator for the
// "Authorization: Bearer <token>" scheme. Verify receives the token with
// the "Bearer " prefix already stripped and returns the authenticated User,
// or ok=false to decline (missing header, malformed, invalid token — this
// Authenticator does not distinguish between them, matching how
// Authenticate never explains a decline to the caller).
type HTTPBearer struct {
	// AuthName overrides the OpenAPI security scheme key (default "bearer").
	AuthName string
	Verify   func(token string) (User, bool)
}

func (a HTTPBearer) Name() string {
	if a.AuthName != "" {
		return a.AuthName
	}
	return "bearer"
}

func (a HTTPBearer) SecurityScheme() openapi.SecurityScheme {
	return openapi.SecurityScheme{Type: "http", Scheme: "bearer"}
}

func (a HTTPBearer) Authenticate(r *http.Request) (User, bool) {
	const prefix = "Bearer "
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, prefix) || a.Verify == nil {
		return nil, false
	}
	token := strings.TrimPrefix(header, prefix)
	if token == "" {
		return nil, false
	}
	return a.Verify(token)
}

// APIKeyHeader is a ready-made Authenticator for a caller-named header
// carrying a credential (default "X-API-Key"). Verify receives the raw
// header value and returns the authenticated User, or ok=false to decline.
type APIKeyHeader struct {
	// HeaderName overrides which header carries the key (default "X-API-Key").
	HeaderName string
	// AuthName overrides the OpenAPI security scheme key (default "apiKey").
	AuthName string
	Verify   func(key string) (User, bool)
}

func (a APIKeyHeader) headerName() string {
	if a.HeaderName != "" {
		return a.HeaderName
	}
	return "X-API-Key"
}

func (a APIKeyHeader) Name() string {
	if a.AuthName != "" {
		return a.AuthName
	}
	return "apiKey"
}

func (a APIKeyHeader) SecurityScheme() openapi.SecurityScheme {
	return openapi.SecurityScheme{Type: "apiKey", In: "header", Name: a.headerName()}
}

func (a APIKeyHeader) Authenticate(r *http.Request) (User, bool) {
	key := r.Header.Get(a.headerName())
	if key == "" || a.Verify == nil {
		return nil, false
	}
	return a.Verify(key)
}

// HTTPBasic is a ready-made Authenticator for RFC 7617 Basic auth
// (an "Authorization: Basic <base64>" header). Verify receives the decoded
// username/password and returns the authenticated User, or ok=false to
// decline.
type HTTPBasic struct {
	// AuthName overrides the OpenAPI security scheme key (default "basic").
	AuthName string
	Verify   func(username, password string) (User, bool)
}

func (a HTTPBasic) Name() string {
	if a.AuthName != "" {
		return a.AuthName
	}
	return "basic"
}

func (a HTTPBasic) SecurityScheme() openapi.SecurityScheme {
	return openapi.SecurityScheme{Type: "http", Scheme: "basic"}
}

func (a HTTPBasic) Authenticate(r *http.Request) (User, bool) {
	username, password, ok := r.BasicAuth()
	if !ok || a.Verify == nil {
		return nil, false
	}
	return a.Verify(username, password)
}

// CookieKey is a ready-made Authenticator for a session token carried in a
// named cookie (default "session"). Verify receives the cookie's raw value
// and returns the authenticated User, or ok=false to decline.
type CookieKey struct {
	// CookieName overrides which cookie carries the token (default "session").
	CookieName string
	// AuthName overrides the OpenAPI security scheme key (default "cookieAuth").
	AuthName string
	Verify   func(value string) (User, bool)
}

func (a CookieKey) cookieName() string {
	if a.CookieName != "" {
		return a.CookieName
	}
	return "session"
}

func (a CookieKey) Name() string {
	if a.AuthName != "" {
		return a.AuthName
	}
	return "cookieAuth"
}

func (a CookieKey) SecurityScheme() openapi.SecurityScheme {
	return openapi.SecurityScheme{Type: "apiKey", In: "cookie", Name: a.cookieName()}
}

func (a CookieKey) Authenticate(r *http.Request) (User, bool) {
	cookie, err := r.Cookie(a.cookieName())
	if err != nil || cookie.Value == "" || a.Verify == nil {
		return nil, false
	}
	return a.Verify(cookie.Value)
}
