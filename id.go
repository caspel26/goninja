package goninja

import "github.com/google/uuid"

// NewUUID returns a new random (v4) UUID as a string. Generated Create()
// methods call this when a model's ID field is typed as a string rather
// than int64 (see Model.IDGoType in internal/codegen) — the framework
// treats a string ID as a UUID primary key it generates itself, since
// unlike an int64 auto-increment column there's no DB-assigned default to
// rely on.
func NewUUID() string {
	return uuid.NewString()
}
