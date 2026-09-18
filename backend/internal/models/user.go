package models

import "time"

type User struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Username       string    `gorm:"size:80;uniqueIndex;not null" json:"username"`
	Name           string    `gorm:"size:120;not null" json:"name"`
	PassphraseHash string    `gorm:"not null" json:"-"`
	Role           Role      `gorm:"size:20;not null;default:reader" json:"role"`
	IsActive       bool      `gorm:"not null;default:true" json:"isActive"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
