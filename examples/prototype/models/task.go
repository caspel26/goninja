// Package models holds the Phase 0 prototype model. It exists to answer
// the goninja-implementation-plan.md Phase 0 decision-gate questions, not
// as a real API — no ORM, one tag, in-memory storage.
package models

// Task is the single model used to validate the generate → use workflow.
type Task struct {
	ID    int64  `json:"id" goninja:"list,retrieve"`
	Title string `json:"title" goninja:"list,retrieve,create"`
	Done  bool   `json:"done" goninja:"list,retrieve,create"`
}
