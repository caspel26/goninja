// WithUser/UserFromContext are the minimal contract between auth
// enforcement and the rest of the framework (plan section 5.8/5.15): an
// Authenticator authenticates the request and returns the resulting User;
// goninja stores it on the context via WithUser before the request reaches
// a resource's handlers, and a resource — most often an overridden method
// or a hook, wired in via BaseResource.SetSelf like everything else in
// Phase 6 — reads it back out via UserFromContext. goninja doesn't impose a
// user struct beyond the one method it needs.
package goninja

import (
	"context"
	"net/http"

	"github.com/caspel26/goninja/openapi"
)

// User is the minimal contract goninja needs from an authenticated caller.
// Implement it on whatever type your own Authenticator already produces —
// goninja never constructs one itself.
type User interface {
	ID() string
}

// Authenticator inspects a request and either identifies the caller or
// declines (plan section 5.15) — mirroring Django Ninja's auth objects
// (HttpBearer, ApiKeyHeader, ...) rather than a plain middleware func: the
// object that enforces auth is also the only source of truth for how it's
// documented, so runtime behavior and the generated OpenAPI document can't
// drift apart the way two separately-wired mechanisms could.
type Authenticator interface {
	// Authenticate inspects r and returns the authenticated User, or
	// ok=false if this Authenticator doesn't recognize the request (e.g. no
	// bearer token present) or the credential is invalid. It never writes
	// to a response itself — goninja rejects the request with 401 only
	// after every configured Authenticator for the route has declined (see
	// BaseResource.Protect, resource.go), which is what lets a route accept
	// more than one scheme (e.g. bearer-or-api-key), tried in order.
	Authenticate(r *http.Request) (User, bool)

	// Name identifies this Authenticator in the generated OpenAPI
	// document — the key under components.securitySchemes and in an
	// Operation's Security requirement. Two distinct Authenticators sharing
	// a Name are assumed to describe the same scheme.
	Name() string

	// SecurityScheme describes this Authenticator for the generated
	// OpenAPI document (see BaseResource.SecurityFor, resource.go).
	SecurityScheme() openapi.SecurityScheme
}

type userKey struct{}

// WithUser returns a context carrying user, so a later UserFromContext call
// anywhere downstream — a hook, an overridden method, an error mapper — can
// retrieve it. Called by your auth middleware once it has authenticated the
// request, before the request reaches a resource's handlers.
func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

// UserFromContext returns the User stored by WithUser, if any. ok is false
// when no middleware set one — e.g. an unauthenticated request, or a route
// not behind auth at all.
func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userKey{}).(User)
	return user, ok
}
