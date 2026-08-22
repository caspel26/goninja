package goninja

import (
	"net/http"

	"github.com/caspel26/goninja/openapi"
)

// Action is a custom endpoint mounted alongside a resource's generated
// CRUD routes — goninja's equivalent of django-ninja-aio-crud's @action.
// Unlike CRUD routes, actions are declared, not generated: call
// BaseResource.SetActions with one or more Actions and the generated
// Register(mux)/OpenAPI() mount and (optionally) document each one
// automatically.
type Action struct {
	// Name identifies the action for ResourceConfig.Auth (keyed by
	// Route(Name)) and Protect, the same way RouteCreate/RouteUpdate/etc
	// identify a CRUD route. Required.
	Name string
	// Detail mounts this action under the resource's item path
	// (<base>/{id}/<UrlPath>) instead of its collection path
	// (<base>/<UrlPath>) — mirrors @action's detail=True/False.
	Detail bool
	// Method is the HTTP method this action responds to, e.g. "POST".
	Method string
	// UrlPath is the path segment appended after the resource's base path
	// (and /{id}, if Detail). May be empty to mount directly on the base
	// (or item) path with a different Method than the generated routes use.
	UrlPath string
	// Handler serves the action.
	Handler http.HandlerFunc
	// Summary documents this action in the resource's OpenAPI fragment.
	// Leave empty to mount the action without documenting it.
	Summary string
	// Responses documents this action's possible responses, keyed by
	// status code. Only used when Summary is set.
	Responses map[string]openapi.Response

	// Auth, if non-nil, decides this action's auth directly — declared
	// alongside the action instead of by re-listing Route(Name) in
	// Config.DefaultAuth.Routes or ResourceConfig.Auth elsewhere, which a
	// new action is easy to add and forget to also list there, leaving it
	// silently public. Public/Auth on the RouteAuth work exactly as they
	// do for ResourceConfig.Auth's entries. Leave nil to fall back to that
	// name-based lookup instead (Route(Name) against ResourceConfig.Auth,
	// then Config.DefaultAuth.Routes) — e.g. when several actions share
	// one app-wide default and repeating it on each isn't worth avoiding
	// the lookup.
	Auth *RouteAuth
}

// Actions returns a constructor Option that builds this resource's
// actions via build(r, arg) and attaches them via SetActions — pass it
// straight to a generated New<Model>Resource call instead of a separate
// SetActions step afterward:
//
//	bookAPI := api.NewBookResource(db, goninja.Actions(bookActions, actionAuth))
//
// where bookActions is func(r *api.BookResource, auth *goninja.RouteAuth)
// []goninja.Action — arg's type isn't fixed to *RouteAuth, it's whatever
// build's second parameter needs. A build func needing nothing beyond r
// doesn't need this at all: New<Model>Resource(db) then r.SetActions(
// bookActions(r)...) is already just as short. The returned func(R) is
// assignable to a generated <Model>Option (an unnamed function value with
// the same underlying signature converts implicitly), which is why R is
// inferred from build rather than needing an explicit type argument at
// the call site above.
func Actions[R interface{ SetActions(...Action) }, A any](build func(R, A) []Action, arg A) func(R) {
	return func(r R) {
		r.SetActions(build(r, arg)...)
	}
}
