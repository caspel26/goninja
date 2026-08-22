package models

import "time"

// Book carries a real belongs-to relation to Author (GORM infers the
// AuthorID/Author pairing by convention, no explicit foreignKey tag
// needed). Only the retrieve schema pulls in the full Author — list stays
// lean by construction. Price/Published/AuthorID/CreatedAt are
// `filter`-tagged, together exercising:
// GET /books?published=true&price_min=10&created_at=2024-01-01T00:00:00Z&order=-created_at&limit=20
// CreatedAt's `filter` tag specifically proves a time.Time exact-match
// filter compiles and works end to end — a real bug found via a
// filter-tagged time.Time field in an external consumer project
// (`?created_at=...` fell through parse<Model>Filters' int64 fallback,
// generating the invalid conversion time.Time(n)) had this exact shape;
// kept tagged permanently as the regression proof, not just a one-off
// repro.
type Book struct {
	ID        string    `gorm:"primaryKey;type:uuid" json:"id" goninja:"list,retrieve"`
	Title     string    `json:"title" goninja:"list,retrieve,create,update" validate:"required,max=200"`
	AuthorID  string    `json:"author_id" goninja:"list,retrieve,create,update,filter" validate:"required,uuid4"`
	Author    Author    `json:"author" goninja:"retrieve"`
	Price     float64   `json:"price" goninja:"list,retrieve,create,update,filter" validate:"min=0"`
	Published bool      `json:"published" goninja:"list,retrieve,create,update,filter"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at" goninja:"list,retrieve,filter"`
}
