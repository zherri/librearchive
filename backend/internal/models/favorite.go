package models

import "time"

type Favorite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_book_favorite" json:"userId"`
	BookID    uint      `gorm:"not null;uniqueIndex:idx_user_book_favorite" json:"bookId"`
	Book      Book      `gorm:"foreignKey:BookID" json:"book,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}
