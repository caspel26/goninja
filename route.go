package goninja

// Route names one of a resource's generated CRUD routes, or a custom
// Action's Name (actions.go) converted via Route(a.Name). Typed rather
// than a bare string so a typo in AuthPolicy.Routes/ResourceConfig.Auth/
// ResourceConfig.Routes is a compile error instead of a route silently
// left unprotected or unrestricted.
type Route string

const (
	RouteList     Route = "list"
	RouteRetrieve Route = "retrieve"
	RouteCreate   Route = "create"
	RouteUpdate   Route = "update"
	RouteDelete   Route = "delete"
)

func containsRoute(list []Route, r Route) bool {
	for _, x := range list {
		if x == r {
			return true
		}
	}
	return false
}
