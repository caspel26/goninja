// Package echoadapter lets goninja's generated resources mount on an echo
// instance or group. Wrap it once with New and pass the result anywhere
// goninja.Router is expected (goninja.API's Mount/MountWithConfig/
// MountDocs, or a resource's own Register).
package echoadapter

import (
	"net/http"

	"github.com/caspel26/goninja/router"
	"github.com/labstack/echo/v4"
)

// Router is the subset of *echo.Echo/*echo.Group's API New needs — both
// satisfy it, so goninja routes can mount at the top level or under an
// existing group/prefix.
type Router interface {
	Add(method, path string, h echo.HandlerFunc, middleware ...echo.MiddlewareFunc) *echo.Route
	Any(path string, h echo.HandlerFunc, middleware ...echo.MiddlewareFunc) []*echo.Route
}

// Adapter satisfies goninja.Router by translating a stdlib-style pattern
// ("GET /books/{id}") into echo's own route registration.
type Adapter struct {
	r Router
}

// New wraps r — a *echo.Echo or any *echo.Group — as a goninja.Router.
func New(r Router) *Adapter {
	return &Adapter{r: r}
}

// HandleFunc registers pattern on the wrapped echo router. It translates
// the pattern's "{name}" wildcards into echo's ":name" syntax and binds
// each matched value back onto the request via SetPathValue before calling
// handler, so a generated handler's req.PathValue("id") call sees the
// right value without any changes to the handler itself.
func (a *Adapter) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	p, err := router.ParsePattern(pattern)
	if err != nil {
		panic("echoadapter: " + err.Error())
	}

	path := p.TranslatePath(router.StyleColon)
	fn := func(c echo.Context) error {
		req := c.Request()
		router.BindPathValues(req, p.Params, c.Param)
		handler(c.Response(), req)
		return nil
	}

	if p.Method == "" {
		a.r.Any(path, fn)
		return
	}
	a.r.Add(p.Method, path, fn)
}
