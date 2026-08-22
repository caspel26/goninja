// Package goninja is the public runtime support package for code generated
// by `goninja generate`: the base resource type generated Resources embed,
// the transaction-aware DB(ctx) contract, and the framework error types.
package goninja

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/caspel26/goninja/openapi"
	"gorm.io/gorm"
)

type txKey struct{}

// WithTx returns a context carrying tx, so that a later BaseResource.DB(ctx)
// call in the same request/transaction returns tx instead of the base
// connection. Used to make hooks and business logic that share a context
// participate in the same transaction.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromContext returns the transaction stored by WithTx, if any.
func TxFromContext(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	return tx, ok
}

// BaseResource is embedded by every generated Resource. It holds the base
// *gorm.DB injected at construction time and exposes it through DB(ctx),
// which is transaction-aware: inside InTransaction, DB(ctx) returns the
// active transaction rather than the base connection.
type BaseResource struct {
	db               *gorm.DB
	errorMapper      ErrorMapper
	openAPITags      []string
	excludedFromDocs bool
	self             any
	config           Config
	actions          []Action
}

// SetDB injects the base database connection. Called by the generated
// New<Model>Resource constructor; not meant to be called directly by users.
func (r *BaseResource) SetDB(db *gorm.DB) {
	r.db = db
}

// SetErrorMapper overrides the ErrorMapper generated handlers use to
// translate resource errors into HTTP responses. Optional — a resource with
// none set falls back to DefaultErrorMapper (see ErrorMapper).
func (r *BaseResource) SetErrorMapper(m ErrorMapper) {
	r.errorMapper = m
}

// ErrorMapper returns the resource's configured ErrorMapper: its own, if
// SetErrorMapper was called; else Config.DefaultErrorMapper, if set (see
// SetConfig); else the package DefaultErrorMapper.
func (r *BaseResource) ErrorMapper() ErrorMapper {
	if r.errorMapper != nil {
		return r.errorMapper
	}
	if r.config.DefaultErrorMapper != nil {
		return r.config.DefaultErrorMapper
	}
	return DefaultErrorMapper{}
}

// SetOpenAPITags overrides the OpenAPI tags every operation this
// resource's generated OpenAPI() puts in its fragment carries — how
// Swagger UI/ReDoc/etc group its routes in a rendered doc. Optional: a resource with none set falls back to a single
// tag equal to the model's name (e.g. "Book"), applied by the generated
// OpenAPI() method itself.
func (r *BaseResource) SetOpenAPITags(tags ...string) {
	r.openAPITags = tags
}

// OpenAPITags returns the resource's configured OpenAPI tags, or nil if
// SetOpenAPITags was never called — generated OpenAPI() methods fall back
// to the model name in that case.
func (r *BaseResource) OpenAPITags() []string {
	return r.openAPITags
}

// ExcludeFromDocs opts this resource out of the OpenAPI document entirely —
// its routes still get mounted by Register(mux) as normal, they just don't
// appear in the merged spec (and so don't show up in Swagger UI/ReDoc/etc).
// Useful for internal or half-built resources you don't want documented
// yet. Mount checks this via the unexported excludedFromDocs so a caller
// doesn't need to route excluded resources by hand.
func (r *BaseResource) ExcludeFromDocs() {
	r.excludedFromDocs = true
}

// DocsExcluded reports whether ExcludeFromDocs was called — checked by
// API.Mount/MountWithConfig via a structural interface.
func (r *BaseResource) DocsExcluded() bool {
	return r.excludedFromDocs
}

// SetSelf tells a generated resource which concrete value the generated
// handlers should dispatch through when checking for overridden methods
// and optional hooks (BeforeCreateHook etc — see hooks.go). Go has no
// dynamic dispatch through embedding: a type that embeds a generated
// <Model>Resource and overrides one of its methods, or implements a hook
// interface on itself, is invisible to code running on the embedded
// receiver unless that receiver is told explicitly.
// The generated New<Model>Resource constructor calls SetSelf(itself) by
// default, so a resource used directly needs no extra step; a type
// embedding it should call SetSelf again in its own constructor, pointing
// it at itself, to make its overrides and hooks take effect.
func (r *BaseResource) SetSelf(self any) {
	r.self = self
}

// Self returns the value passed to SetSelf, or nil if it was never called.
func (r *BaseResource) Self() any {
	return r.self
}

// SetConfig injects the app-wide Config (global default auth, generic
// middleware, default error mapper) this resource's generated
// Register(mux) uses to build its handlers. Called by MountWithConfig
// always, and by Mount too when API.SetErrorMapper was used — otherwise a
// resource mounted via plain Mount never gets this called, so Config()
// stays at its zero value: no global auth, no global middleware, no
// default error mapper.
func (r *BaseResource) SetConfig(cfg Config) {
	r.config = cfg
}

// Config returns the resource's configured Config, set via SetConfig.
func (r *BaseResource) Config() Config {
	return r.config
}

// SetActions declares custom endpoints beyond the generated CRUD set — see
// Action. The generated Register(mux) mounts each one automatically
// (wrapped through Protect, same as the CRUD routes) and OpenAPI()
// documents every Action with a Summary set. No wrapper type or SetSelf
// needed just for this: unlike hooks/Configurer, an Action carries its own
// http.HandlerFunc, so there's no per-request dispatch to resolve — a
// resource used directly can call SetActions right after construction.
func (r *BaseResource) SetActions(actions ...Action) {
	r.actions = actions
}

// Actions returns the resource's declared Actions, set via SetActions, or
// nil if none were declared.
func (r *BaseResource) Actions() []Action {
	return r.actions
}

// Protect wraps h according to this resource's global Config
// (DefaultAuth/Middleware, config.go) and rc's per-resource RouteAuth
// override (resource_config.go): route is protected
// when a rc.Auth[route] override says Public=false and (its own Auth or,
// if unset, Config.DefaultAuth.Auth) applies, or — absent an override —
// when route is named in Config.DefaultAuth.Routes. A protected route is
// wrapped so any one of its resolved Authenticators recognizing the
// request lets it through (tried in order), stores the resulting User via
// WithUser, and otherwise responds 401 once every Authenticator has
// declined. Config.Middleware always wraps h, protected or not. Generated
// Register(mux) methods call this around every CRUD handler they mount;
// ProtectAction is the equivalent for an Action.
func (r *BaseResource) Protect(route Route, rc ResourceConfig, h http.HandlerFunc) http.HandlerFunc {
	auths, protected := r.authenticatorsFor(route, rc)
	return r.protect(h, auths, protected)
}

// ProtectAction is Protect for an Action: consults a.Auth first (see
// Action.Auth), falling back to the same Route(a.Name)-keyed lookup
// Protect itself uses for CRUD routes when a.Auth is nil. Generated
// Register(mux) methods call this around every Action they mount.
func (r *BaseResource) ProtectAction(a Action, rc ResourceConfig) http.HandlerFunc {
	auths, protected := r.authenticatorsForAction(a, rc)
	return r.protect(a.Handler, auths, protected)
}

// protect is Protect/ProtectAction's shared tail once each has resolved
// its own (auths, protected) pair — see authenticatorsFor/
// authenticatorsForAction.
func (r *BaseResource) protect(h http.HandlerFunc, auths []Authenticator, protected bool) http.HandlerFunc {
	wrapped := http.Handler(h)
	if protected {
		wrapped = requireAuth(auths, wrapped)
	}
	for i := len(r.config.Middleware) - 1; i >= 0; i-- {
		wrapped = r.config.Middleware[i](wrapped)
	}
	return wrapped.ServeHTTP
}

// SecurityFor mirrors Protect's own protected/override resolution so
// generated OpenAPI() methods document exactly what Protect enforces,
// without re-implementing that logic: reqs is an OpenAPI Security
// requirement list (nil if route isn't protected, one alternative entry
// per resolved Authenticator), and schemes collects each Authenticator's
// SecurityScheme keyed by Name for Components.SecuritySchemes.
// SecurityForAction is the equivalent for an Action.
func (r *BaseResource) SecurityFor(route Route, rc ResourceConfig) (reqs []map[string][]string, schemes map[string]openapi.SecurityScheme) {
	auths, protected := r.authenticatorsFor(route, rc)
	return r.securityFor(auths, protected)
}

func (r *BaseResource) SecurityForAction(a Action, rc ResourceConfig) (reqs []map[string][]string, schemes map[string]openapi.SecurityScheme) {
	auths, protected := r.authenticatorsForAction(a, rc)
	return r.securityFor(auths, protected)
}

func (r *BaseResource) securityFor(auths []Authenticator, protected bool) (reqs []map[string][]string, schemes map[string]openapi.SecurityScheme) {
	if !protected || len(auths) == 0 {
		return nil, nil
	}
	schemes = make(map[string]openapi.SecurityScheme, len(auths))
	for _, a := range auths {
		reqs = append(reqs, map[string][]string{a.Name(): {}})
		schemes[a.Name()] = a.SecurityScheme()
	}
	return reqs, schemes
}

func (r *BaseResource) authenticatorsFor(route Route, rc ResourceConfig) (auths []Authenticator, protected bool) {
	if ra, ok := rc.Auth[route]; ok {
		return r.resolveRouteAuth(ra)
	}
	if containsRoute(r.config.DefaultAuth.Routes, route) {
		return r.config.DefaultAuth.Auth, true
	}
	return nil, false
}

// authenticatorsForAction is authenticatorsFor's Action-aware counterpart:
// a.Auth, if set, decides the outcome directly (see Action.Auth); nil
// falls back to the same Route(a.Name)-keyed lookup a CRUD route uses.
func (r *BaseResource) authenticatorsForAction(a Action, rc ResourceConfig) (auths []Authenticator, protected bool) {
	if a.Auth != nil {
		return r.resolveRouteAuth(*a.Auth)
	}
	return r.authenticatorsFor(Route(a.Name), rc)
}

// resolveRouteAuth is the RouteAuth -> (auths, protected) resolution
// shared by an explicit ResourceConfig.Auth[route] entry and an explicit
// Action.Auth: Public wins outright; a non-empty Auth replaces the
// default; an empty RouteAuth{} opts into Config.DefaultAuth.Auth.
func (r *BaseResource) resolveRouteAuth(ra RouteAuth) (auths []Authenticator, protected bool) {
	if ra.Public {
		return nil, false
	}
	if len(ra.Auth) > 0 {
		return ra.Auth, true
	}
	return r.config.DefaultAuth.Auth, true
}

// classified reports whether route has an explicit auth decision anywhere
// — an entry in rc.Auth, or named in Config.DefaultAuth.Routes — distinct
// from whether it's actually protected: RouteAuth{Public: true} is also
// an explicit decision, so a route left public on purpose doesn't count
// as unclassified. Used by CheckStrictAuth to catch the route nobody
// actually decided about, not the one deliberately left open.
func (r *BaseResource) classified(route Route, rc ResourceConfig) bool {
	if _, ok := rc.Auth[route]; ok {
		return true
	}
	return containsRoute(r.config.DefaultAuth.Routes, route)
}

// classifiedAction is classified's Action-aware counterpart: a's own Auth,
// if set, is itself the explicit decision (see Action.Auth); nil falls
// back to classified's Route(a.Name)-keyed check.
func (r *BaseResource) classifiedAction(a Action, rc ResourceConfig) bool {
	if a.Auth != nil {
		return true
	}
	return r.classified(Route(a.Name), rc)
}

// CheckStrictAuth panics if Config.StrictAuth is set and any route in
// routes, or any Action in actions, has no explicit auth decision (see
// classified/classifiedAction) — a no-op when StrictAuth is false (the
// default), so an app that never opts in sees no behavior change.
// Generated Register(mux) methods call this once, with every route
// they're about to mount, before mounting any of them — so a startup
// panic replaces what would otherwise be a silently public endpoint,
// rather than that endpoint working normally until someone notices.
func (r *BaseResource) CheckStrictAuth(routes []Route, actions []Action, rc ResourceConfig) {
	if !r.config.StrictAuth {
		return
	}
	var unclassified []string
	for _, route := range routes {
		if !r.classified(route, rc) {
			unclassified = append(unclassified, string(route))
		}
	}
	for _, a := range actions {
		if !r.classifiedAction(a, rc) {
			unclassified = append(unclassified, a.Name)
		}
	}
	if len(unclassified) > 0 {
		panic(fmt.Sprintf(
			"goninja: StrictAuth is set but these routes have no explicit auth decision: %s — "+
				"add each to ResourceConfig.Auth or Action.Auth (RouteAuth{Public: true} if intentional), "+
				"or name it in Config.DefaultAuth.Routes",
			strings.Join(unclassified, ", "),
		))
	}
}

// requireAuth wraps next so the request reaches it only once one of auths
// recognizes it — tried in order, first match wins, its User attached to
// the request context via WithUser. Responds 401 if every Authenticator
// declines.
func requireAuth(auths []Authenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, a := range auths {
			if user, ok := a.Authenticate(req); ok {
				next.ServeHTTP(w, req.WithContext(WithUser(req.Context(), user)))
				return
			}
		}
		Respond(w, nil, Unauthorized{})
	})
}

// DB returns the connection to use for this context: the enclosing
// transaction if ctx carries one (see WithTx/InTransaction), otherwise the
// base connection bound with ctx.
func (r *BaseResource) DB(ctx context.Context) *gorm.DB {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

// InTransaction runs fn inside a database transaction, exposing it to fn
// via ctx so that any BaseResource.DB(ctx) call made from within fn — by
// the resource itself or by a hook — participates in the same transaction.
func InTransaction[T any](ctx context.Context, db *gorm.DB, fn func(ctx context.Context) (T, error)) (T, error) {
	var result T
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		out, err := fn(WithTx(ctx, tx))
		if err != nil {
			return err
		}
		result = out
		return nil
	})
	return result, err
}
