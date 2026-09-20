package models

import "time"

type Highlight struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null;index" json:"userId"`
	BookID       uint      `gorm:"not null;index" json:"bookId"`
	SelectedText string    `gorm:"type:text;not null" json:"selectedText"`
	Color        string    `gorm:"size:32;not null;default:yellow;index" json:"color"`
	Page         int       `gorm:"not null" json:"page"`
	StartOffset  *int      `json:"startOffset,omitempty"`
	EndOffset    *int      `json:"endOffset,omitempty"`
	ChapterRef   string    `gorm:"size:255" json:"chapterRef,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Note         *Note     `gorm:"foreignKey:HighlightID" json:"note,omitempty"`
}
