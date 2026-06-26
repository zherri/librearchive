package models

type CollectionBook struct {
	ID           uint `gorm:"primaryKey" json:"collectionbook_id"`
	CollectionID uint `gorm:"not null;" json:"collection_id"`
	BookID       uint `gorm:"not null;" json:"book_id"`
}
