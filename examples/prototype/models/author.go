package models

// Author is the second prototype model, distinct in shape from Task, used
// to prove the generator isn't special-cased to a single struct (Phase 1
// exit criterion). Book (below) references it as a relation, proving
// automatic Preload on Retrieve (Phase 2 exit criterion).
type Author struct {
	ID   int64  `gorm:"primaryKey" json:"id" goninja:"list,retrieve"`
	Name string `json:"name" goninja:"list,retrieve,create,update" validate:"required,max=120"`
	Bio  string `json:"bio" goninja:"retrieve,create,update" validate:"max=2000"`
}
