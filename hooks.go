// Hooks are the optional interfaces a resource can implement to run logic
// around Create/Update/Delete without touching generated code. None of
// these are implemented by a generated
// <Model>Resource itself — implement them on a type that embeds one and is
// wired in via BaseResource.SetSelf, so the generated handlers can see it
// (see resource.go's Self/SetSelf).
//
// Every hook runs inside the same database transaction as the operation it
// wraps (goninja.InTransaction, resource.go): returning an error from a
// hook aborts the whole request and rolls back anything the operation
// itself already did.
package goninja

import "context"

// BeforeCreateHook runs before Create writes its row. TIn is the model's
// generated Create schema (e.g. BookCreate) — implement
// BeforeCreate(ctx, *BookCreate) error to mutate or validate the input
// before it's persisted.
type BeforeCreateHook[TIn any] interface {
	BeforeCreate(ctx context.Context, in *TIn) error
}

// AfterCreateHook runs after Create has written and reloaded the row, still
// inside the same transaction. TOut is the model's generated Retrieve
// schema (e.g. BookRetrieve).
type AfterCreateHook[TOut any] interface {
	AfterCreate(ctx context.Context, out *TOut) error
}

// BeforeUpdateHook mirrors BeforeCreateHook for Update. TIn is the model's
// generated Update schema (e.g. BookUpdate).
type BeforeUpdateHook[TIn any] interface {
	BeforeUpdate(ctx context.Context, in *TIn) error
}

// BeforeDeleteHook runs before Delete removes the row. TID is the model's
// ID type (int64 or string — see Model.IDGoType).
type BeforeDeleteHook[TID any] interface {
	BeforeDelete(ctx context.Context, id TID) error
}
