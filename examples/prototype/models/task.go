// Package models holds the prototype models exercising the goninja
// generator end to end — see goninja-implementation-plan.md phases 0-2.
package models

// Task is the first prototype model.
type Task struct {
	ID    int64  `gorm:"primaryKey" json:"id" goninja:"list,retrieve"`
	Title string `json:"title" goninja:"list,retrieve,create,update"`
	Done  bool   `json:"done" goninja:"list,retrieve,create,update"`
}
