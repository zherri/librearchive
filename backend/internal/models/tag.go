package models

import "time"

type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:80;not null;uniqueIndex" json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type BookTag struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	BookID uint `gorm:"not null;uniqueIndex:idx_book_tag" json:"bookId"`
	TagID  uint `gorm:"not null;uniqueIndex:idx_book_tag" json:"tagId"`
}
