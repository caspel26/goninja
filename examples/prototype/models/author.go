package models

// Author is the second prototype model, distinct in shape from Task, used
// to prove the generator isn't special-cased to a single struct.
// Book (below) references it as a belongs-to relation,
// proving automatic Preload on Retrieve; Books
// here is the reverse side — a has-many relation (GORM infers it from
// Book.AuthorID by convention, same as the belongs-to side) — proving
// codegen's has-many support (Field.IsSlice, internal/codegen/ir.go).
type Author struct {
	ID    string `gorm:"primaryKey;type:uuid" json:"id" goninja:"list,retrieve"`
	Name  string `json:"name" goninja:"list,retrieve,create,update,filter" validate:"required,max=120"`
	Bio   string `json:"bio" goninja:"retrieve,create,update" validate:"max=2000"`
	Books []Book `json:"books" goninja:"retrieve"`
}
