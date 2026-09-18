package models

import "time"

type AuthSession struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	UserID           uint       `gorm:"not null;index" json:"userId"`
	RefreshTokenHash string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	ExpiresAt        time.Time  `gorm:"not null;index" json:"expiresAt"`
	RevokedAt        *time.Time `json:"revokedAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
}
