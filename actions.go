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
	// Name identifies the action for ResourceConfig.Auth
	// (AlsoProtect/Public) and Protect, the same way "list"/"create"/etc
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
}
