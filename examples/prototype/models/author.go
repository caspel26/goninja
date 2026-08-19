package models

// Author is the second prototype model, distinct in shape from Task
// (different field count/types), used to prove the generator isn't
// accidentally special-cased to a single struct — Phase 1 exit criterion
// in goninja-implementation-plan.md.
type Author struct {
	ID        int64  `json:"id" goninja:"list,retrieve"`
	Name      string `json:"name" goninja:"list,retrieve,create"`
	Bio       string `json:"bio" goninja:"retrieve,create"`
	BookCount int64  `json:"book_count" goninja:"list,retrieve"`
}
