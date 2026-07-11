package models

import (
	"time"
)

type User struct {
	ID               uint      `gorm:"primaryKey" json:"user_id"`
	Username         string    `gorm:"type:varchar(80);unique;not null" json:"username"`
	Passphrase       string    `gorm:"type:varchar(255);not null" json:"-"`
	RegistrationDate time.Time `gorm:"type:timestamptz;autoCreateTime" json:"registration_date"`

	BookProgressions []BookProgress `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"progressions,omitempty"`
	Annotations      []Annotation   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"annotations,omitempty"`
	Collections      []Collection   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"collections,omitempty"`
}
