// Package goninja is the public runtime support package for code generated
// by `goninja generate`: the base resource type generated Resources embed,
// the transaction-aware DB(ctx) contract, and the framework error types.
//
// See goninja-implementation-plan.md for the full design; this file covers
// the Phase 2 slice of it (GORM + transaction-aware queries).
package goninja

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// WithTx returns a context carrying tx, so that a later BaseResource.DB(ctx)
// call in the same request/transaction returns tx instead of the base
// connection. Used to make hooks and business logic that share a context
// participate in the same transaction (plan section 5.7).
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

// ErrorMapper returns the resource's configured ErrorMapper, or
// DefaultErrorMapper if none was set via SetErrorMapper.
func (r *BaseResource) ErrorMapper() ErrorMapper {
	if r.errorMapper == nil {
		return DefaultErrorMapper{}
	}
	return r.errorMapper
}

// SetOpenAPITags overrides the OpenAPI tags every operation this
// resource's generated OpenAPI() puts in its fragment carries — how
// Swagger UI/ReDoc/etc group its routes in a rendered doc (plan section
// 5.10/Fase 5). Optional: a resource with none set falls back to a single
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

func (r *BaseResource) docsExcluded() bool {
	return r.excludedFromDocs
}

// SetSelf tells a generated resource which concrete value the generated
// handlers should dispatch through when checking for overridden methods
// and optional hooks (BeforeCreateHook etc — see hooks.go). Go has no
// dynamic dispatch through embedding: a type that embeds a generated
// <Model>Resource and overrides one of its methods, or implements a hook
// interface on itself, is invisible to code running on the embedded
// receiver unless that receiver is told explicitly (plan section 5.10).
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
