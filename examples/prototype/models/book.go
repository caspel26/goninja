package models

import "time"

// Book carries a real belongs-to relation to Author (GORM infers the
// AuthorID/Author pairing by convention, no explicit foreignKey tag
// needed). Only the retrieve schema pulls in the full Author — list stays
// lean by construction, per plan section 5.5. Price/Published/AuthorID are
// `filter`-tagged and CreatedAt is orderable, together exercising the
// Phase 4 exit criterion:
// GET /books?published=true&price_min=10&order=-created_at&limit=20
type Book struct {
	ID        string    `gorm:"primaryKey;type:uuid" json:"id" goninja:"list,retrieve"`
	Title     string    `json:"title" goninja:"list,retrieve,create,update" validate:"required,max=200"`
	AuthorID  string    `json:"author_id" goninja:"list,retrieve,create,update,filter" validate:"required,uuid4"`
	Author    Author    `json:"author" goninja:"retrieve"`
	Price     float64   `json:"price" goninja:"list,retrieve,create,update,filter" validate:"min=0"`
	Published bool      `json:"published" goninja:"list,retrieve,create,update,filter"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at" goninja:"list,retrieve"`
}
