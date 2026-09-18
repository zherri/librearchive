package models

import "time"

type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:120;not null;uniqueIndex" json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type BookCategory struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	BookID     uint `gorm:"not null;uniqueIndex:idx_book_category" json:"bookId"`
	CategoryID uint `gorm:"not null;uniqueIndex:idx_book_category" json:"categoryId"`
}
