package models

import "time"

type Book struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Title          string    `gorm:"size:255;not null;index" json:"title"`
	Authors        string    `gorm:"size:1000;not null;default:'';index" json:"authors"`
	Description    string    `gorm:"type:text" json:"description"`
	Publisher      string    `gorm:"size:255" json:"publisher"`
	PublishedYear  *int      `json:"publishedYear,omitempty"`
	PageCount      *int      `json:"pageCount,omitempty"`
	StoredFilename string    `gorm:"size:255;not null;uniqueIndex" json:"-"`
	OriginalName   string    `gorm:"size:255;not null" json:"originalName"`
	CoverFilename  string    `gorm:"size:255" json:"-"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
