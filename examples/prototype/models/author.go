package models

import "time"

// Author is the second prototype model, distinct in shape from Task, used
// to prove the generator isn't special-cased to a single struct.
// Book (below) references it as a belongs-to relation,
// proving automatic Preload on Retrieve; Books
// here is the reverse side — a has-many relation (GORM infers it from
// Book.AuthorID by convention, same as the belongs-to side) — proving
// codegen's has-many support (Field.IsSlice, internal/codegen/ir.go).
//
// CreatedAt is deliberately the only non-string `filter` field here, which
// makes Author the regression proof for a v0.5.0 bug an external consumer
// found: Model.NeedsStrconv counted any non-string filter field as needing
// the "strconv" import, but a time.Time one parses with time.Parse and
// never touches strconv — so this exact shape (string filters plus a
// time.Time one, no bool/int/float) generated an unused import that failed
// to compile. Book can't prove it: its float/bool filters use strconv
// legitimately and mask the bug.
type Author struct {
	ID        string    `gorm:"primaryKey;type:uuid" json:"id" goninja:"list,retrieve"`
	Name      string    `json:"name" goninja:"list,retrieve,create,update,filter" validate:"required,max=120"`
	Bio       string    `json:"bio" goninja:"retrieve,create,update" validate:"max=2000"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_on" json:"created_at" goninja:"list,retrieve,filter"`
	Books     []Book    `json:"books" goninja:"retrieve"`
}
