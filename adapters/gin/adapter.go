// Package ginadapter lets goninja's generated resources mount on a gin
// engine or router group. Wrap it once with New and pass the result
// anywhere goninja.Router is expected (goninja.API's Mount/MountWithConfig/
// MountDocs, or a resource's own Register).
package ginadapter

import (
	"net/http"

	"github.com/caspel26/goninja/router"
	"github.com/gin-gonic/gin"
)

// Adapter satisfies goninja.Router by translating a stdlib-style pattern
// ("GET /books/{id}") into gin's own route registration.
type Adapter struct {
	r gin.IRouter
}

// New wraps r — a *gin.Engine or any *gin.RouterGroup (so goninja routes
// can be mounted under an existing prefix and behind existing gin
// middleware) — as a goninja.Router.
func New(r gin.IRouter) *Adapter {
	return &Adapter{r: r}
}

// HandleFunc registers pattern on the wrapped gin router. It translates
// the pattern's "{name}" wildcards into gin's ":name" syntax and binds
// each matched value back onto the request via SetPathValue before calling
// handler, so a generated handler's req.PathValue("id") call sees the
// right value without any changes to the handler itself.
func (a *Adapter) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	p, err := router.ParsePattern(pattern)
	if err != nil {
		panic("ginadapter: " + err.Error())
	}

	path := p.TranslatePath(router.StyleColon)
	fn := func(c *gin.Context) {
		router.BindPathValues(c.Request, p.Params, c.Param)
		handler(c.Writer, c.Request)
	}

	if p.Method == "" {
		a.r.Any(path, fn)
		return
	}
	a.r.Handle(p.Method, path, fn)
}
