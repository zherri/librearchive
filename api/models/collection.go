package models

type Collection struct {
	ID     uint   `gorm:"primaryKey" json:"collection_id"`
	UserID uint   `gorm:"not null;" json:"user_id"`
	Name   string `gorm:"type:varchar(100);not null" json:"name"`

	Books []CollectionBook `gorm:"foreignKey:CollectionID;constraint:OnDelete:CASCADE" json:"books,omitempty"`
}
