package models

import "time"

type Collection struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	Name      string    `gorm:"size:120;not null" json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CollectionBook struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CollectionID uint      `gorm:"not null;uniqueIndex:idx_collection_book" json:"collectionId"`
	BookID       uint      `gorm:"not null;uniqueIndex:idx_collection_book" json:"bookId"`
	Book         Book      `gorm:"foreignKey:BookID" json:"book,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}
