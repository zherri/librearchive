package models

import "time"

type Book struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Title          string    `gorm:"size:255;not null;index" json:"title"`
	Author         string    `gorm:"size:255;not null;index" json:"author"`
	Description    string    `gorm:"type:text" json:"description"`
	Language       string    `gorm:"size:20" json:"language"`
	PublishedYear  *int      `json:"publishedYear,omitempty"`
	PageCount      *int      `json:"pageCount,omitempty"`
	StoredFilename string    `gorm:"size:255;not null;uniqueIndex" json:"-"`
	OriginalName   string    `gorm:"size:255;not null" json:"originalName"`
	CoverFilename  string    `gorm:"size:255" json:"-"`
	UploadedByID   uint      `gorm:"not null;index" json:"uploadedById"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
