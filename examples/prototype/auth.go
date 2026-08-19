package main

import (
	"crypto/subtle"

	"github.com/caspel26/goninja"
)

type apiKeyUser struct{}

func (apiKeyUser) ID() string { return "api-key-client" }

// newAPIKeyAuth builds a goninja.APIKeyHeader — the built-in Authenticator
// for a credential carried in a header — proving it end to end against a
// real generated resource; see main.go's wiring. A real deployment would
// look the key up against a store instead of comparing against one fixed
// value, but the constant-time comparison here still matters either way.
func newAPIKeyAuth(key string) goninja.APIKeyHeader {
	return goninja.APIKeyHeader{
		Verify: func(got string) (goninja.User, bool) {
			if subtle.ConstantTimeCompare([]byte(got), []byte(key)) != 1 {
				return nil, false
			}
			return apiKeyUser{}, true
		},
	}
}
