// Package models holds the single model this example generates a resource
// for — see examples/prototype/models for a fuller model set (relations,
// filters, byid) exercised against the stdlib net/http path.
package models

// Task is a minimal model, just enough to prove a generated resource
// mounts correctly on gin via adapters/gin.
type Task struct {
	ID    string `gorm:"primaryKey;type:uuid" json:"id" goninja:"list,retrieve"`
	Title string `json:"title" goninja:"list,retrieve,create,update" validate:"required,max=200"`
	Done  bool   `json:"done" goninja:"list,retrieve,create,update,filter"`
}
