// WithUser/UserFromContext are the minimal contract between an auth
// middleware (yours) and the framework (plan section 5.8, Fase 6 item 4):
// middleware authenticates the request and stores the resulting User on the
// context; a resource — most often an overridden method or a hook, wired in
// via BaseResource.SetSelf like everything else in Phase 6 — reads it back
// out. goninja doesn't impose a user struct beyond the one method it needs.
//
// Nothing here enforces authentication yet — that's Config.DefaultAuth/
// Config.Middleware (plan section 6 item 5), still to come. This file only
// carries the user through the request; a resource that wants to require
// one checks UserFromContext itself in the meantime.
package mw

import "context"

// User is the minimal contract goninja needs from an authenticated caller.
// Implement it on whatever type your own auth middleware already produces —
// goninja never constructs one itself.
type User interface {
	ID() string
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
