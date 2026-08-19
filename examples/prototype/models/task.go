// Package models holds the prototype models exercising the goninja
// generator end to end — see goninja-implementation-plan.md phases 0-4.
package models

// Task is the first prototype model. Its ID is a string (UUID) primary
// key rather than an int64 auto-increment column — goninja.NewUUID fills
// it in on Create — proving the generator's ID type isn't hardcoded to
// int64 (see Model.IDGoType in internal/codegen/ir.go).
type Task struct {
	ID    string `gorm:"primaryKey;type:uuid" json:"id" goninja:"list,retrieve"`
	Title string `json:"title" goninja:"list,retrieve,create,update" validate:"required,max=200"`
	Done  bool   `json:"done" goninja:"list,retrieve,create,update,filter"`
}
