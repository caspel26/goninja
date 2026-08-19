package models

// Book carries a real belongs-to relation to Author (GORM infers the
// AuthorID/Author pairing by convention, no explicit foreignKey tag
// needed). Only the retrieve schema pulls in the full Author — list stays
// lean by construction, per plan section 5.5.
type Book struct {
	ID       int64  `gorm:"primaryKey" json:"id" goninja:"list,retrieve"`
	Title    string `json:"title" goninja:"list,retrieve,create,update"`
	AuthorID int64  `json:"author_id" goninja:"list,retrieve,create,update"`
	Author   Author `json:"author" goninja:"retrieve"`
}
