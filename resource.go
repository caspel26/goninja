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
	db *gorm.DB
}

// SetDB injects the base database connection. Called by the generated
// New<Model>Resource constructor; not meant to be called directly by users.
func (r *BaseResource) SetDB(db *gorm.DB) {
	r.db = db
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
