// Package router is the shared vocabulary between goninja's generated
// Register methods and any HTTP router that wants to host them. *http.ServeMux
// already satisfies Router as-is; adapters for other routers (gin, echo,
// chi, ...) translate a stdlib-style pattern into their own syntax and bind
// path parameters back onto *http.Request via SetPathValue, so a generated
// handler written against req.PathValue("id") never has to know which
// router matched the request.
package router

import (
	"fmt"
	"net/http"
	"strings"
)

// Router is what a generated Register method mounts routes onto.
// *http.ServeMux satisfies it without any wrapping.
type Router interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

// Pattern is a parsed net/http.ServeMux-style pattern, e.g. "GET /books/{id}".
type Pattern struct {
	// Method is the HTTP method the pattern was registered for, or "" if
	// the pattern carried no method (matches any method).
	Method string
	// Path is the pattern's path, e.g. "/books/{id}".
	Path string
	// Params are the named wildcards in Path, in order, e.g. ["id"].
	Params []string
	// Wildcard is the trailing "{name...}" segment's name, or "" if Path
	// has no trailing wildcard.
	Wildcard string
	// Subtree is true when Path ends in "/", meaning ServeMux would treat
	// it as a subtree match rather than an exact one.
	Subtree bool
}

// ParsePattern parses a net/http.ServeMux-style pattern into its method and
// path components, and extracts every named wildcard. It rejects
// host-prefixed patterns ("example.com/x"), which ServeMux supports but no
// goninja-generated pattern ever uses.
func ParsePattern(pattern string) (Pattern, error) {
	method, rest, hasMethod := strings.Cut(pattern, " ")
	if !hasMethod {
		rest = method
		method = ""
	}
	if rest == "" {
		return Pattern{}, fmt.Errorf("router: empty path in pattern %q", pattern)
	}
	if rest[0] != '/' {
		return Pattern{}, fmt.Errorf("router: host-prefixed patterns are not supported: %q", pattern)
	}

	p := Pattern{Method: method, Path: rest, Subtree: strings.HasSuffix(rest, "/")}

	segments := strings.Split(rest, "/")
	for _, seg := range segments {
		if len(seg) < 2 || seg[0] != '{' || seg[len(seg)-1] != '}' {
			continue
		}
		name := seg[1 : len(seg)-1]
		if wc, ok := strings.CutSuffix(name, "..."); ok {
			p.Wildcard = wc
			continue
		}
		p.Params = append(p.Params, name)
	}

	return p, nil
}

// ParamStyle selects the path-parameter syntax TranslatePath rewrites a
// Pattern's Path into.
type ParamStyle int

const (
	// StyleBrace is stdlib/chi's "{name}" syntax — TranslatePath is a
	// no-op for a Pattern already in this style.
	StyleBrace ParamStyle = iota
	// StyleColon is gin/echo's ":name" (and "*name" for a trailing
	// wildcard) syntax.
	StyleColon
)

// TranslatePath rewrites p.Path's "{name}"/"{name...}" wildcards into
// style's syntax. Non-wildcard segments are left untouched.
func (p Pattern) TranslatePath(style ParamStyle) string {
	if style == StyleBrace {
		return p.Path
	}

	segments := strings.Split(p.Path, "/")
	for i, seg := range segments {
		if len(seg) < 2 || seg[0] != '{' || seg[len(seg)-1] != '}' {
			continue
		}
		name := seg[1 : len(seg)-1]
		if wc, ok := strings.CutSuffix(name, "..."); ok {
			segments[i] = "*" + wc
			continue
		}
		segments[i] = ":" + name
	}
	return strings.Join(segments, "/")
}

// BindPathValues copies each of names' values (read from get) onto req via
// SetPathValue, so a handler written against req.PathValue(name) sees the
// right value regardless of which router matched the request.
func BindPathValues(req *http.Request, names []string, get func(string) string) {
	for _, name := range names {
		req.SetPathValue(name, get(name))
	}
}
