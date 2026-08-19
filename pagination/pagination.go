// Package pagination is Fase 7's split of goninja's limit/offset parsing
// and ListEnvelope wrapper out of the root goninja package — it depends
// only on goninja itself, for BadRequest.
package pagination

import (
	"net/url"
	"strconv"

	"github.com/caspel26/goninja"
)

// DefaultLimit is the page size generated List handlers use when the
// request supplies no "limit" query parameter.
const DefaultLimit = 20

// MaxLimit caps how large a "limit" query parameter is allowed to be,
// regardless of what the caller asks for.
const MaxLimit = 100

// ParseLimitOffset parses the "limit"/"offset" query parameters shared by
// every generated List handler (plan section 5.9/Fase 4), applying
// DefaultLimit/MaxLimit bounds. A non-integer or negative value is a
// BadRequest, not silently ignored.
func ParseLimitOffset(q url.Values) (limit, offset int, err error) {
	limit = DefaultLimit
	if v := q.Get("limit"); v != "" {
		n, convErr := strconv.Atoi(v)
		if convErr != nil || n < 0 {
			return 0, 0, goninja.BadRequest{Detail: "invalid limit"}
		}
		limit = n
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	if v := q.Get("offset"); v != "" {
		n, convErr := strconv.Atoi(v)
		if convErr != nil || n < 0 {
			return 0, 0, goninja.BadRequest{Detail: "invalid offset"}
		}
		offset = n
	}

	return limit, offset, nil
}

// ListEnvelope is the response shape every generated List handler wraps
// its page in (plan section 5.9) — Retrieve/Create/Update return the bare
// object, List always returns this envelope so total/limit/offset travel
// alongside the page without the caller needing a second request.
type ListEnvelope[T any] struct {
	Items  []T   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}
