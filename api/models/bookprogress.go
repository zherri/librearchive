package models

import "time"

type BookProgress struct {
	ID           uint      `gorm:"primaryKey" json:"bookprogress_id"`
	UserID       uint      `gorm:"not null;" json:"user_id"`
	BookID       uint      `gorm:"not null;" json:"book_id"`
	LastCFI      string    `gorm:"type:varchar(255)" json:"last_cfi,omitempty"`
	LastPage     int       `json:"last_page,omitempty"`
	Percentage   float64   `gorm:"type:decimal(5,2)" json:"percentage"`
	Status       string    `gorm:"type:varchar(20);default:'reading'" json:"status"`
	IsFavorite   bool      `gorm:"default:false" json:"is_favorite"`
	IsDownloaded bool      `gorm:"default:false" json:"is_downloaded"`
	UpdatedAt    time.Time `json:"updated_at"`
}
