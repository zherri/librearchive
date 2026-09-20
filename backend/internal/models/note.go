package models

import "time"

type Note struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"userId"`
	BookID      uint      `gorm:"not null;index" json:"bookId"`
	HighlightID uint      `gorm:"not null;uniqueIndex" json:"highlightId"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
