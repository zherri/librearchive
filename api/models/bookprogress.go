package models

import "time"

type BookProgress struct {
	ID           uint      `gorm:"primaryKey" json:"bookprogress_id"`
	UserID       uint      `gorm:"not null;" json:"user_id"`
	BookID       uint      `gorm:"not null;" json:"book_id"`
	LastCFI      *string   `gorm:"type:varchar(255)" json:"last_cfi,omitempty"`
	LastPage     *int      `json:"last_page,omitempty"`
	Percentage   float64   `gorm:"type:decimal(5,2)" json:"percentage"`
	Rating       *float32  `gorm:"type:decimal(3,2)" json:"rating"`
	Review       *string   `gorm:"type:text" json:"review"`
	Status       string    `gorm:"type:varchar(20);default:'reading'" json:"status"`
	IsFavorite   bool      `gorm:"default:false" json:"is_favorite"`
	IsDownloaded bool      `gorm:"default:false" json:"is_downloaded"`
	UpdatedAt    time.Time `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Book *Book `gorm:"foreignKey:BookID" json:"book,omitempty"`
}
