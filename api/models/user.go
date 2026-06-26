package models

import (
	"time"
)

type User struct {
	ID               uint      `gorm:"primaryKey" json:"user_id"`
	Username         string    `gorm:"type:varchar(80);unique;not null" json:"username"`
	Password         string    `gorm:"type:varchar(255);not null" json:"-"`
	RegistrationDate time.Time `gorm:"type:timestamptz;autoCreateTime" json:"registration_date"`

	BookProgressions []BookProgress `gorm:"foreignKey:UserID" json:"progressions,omitempty"`
	Annotations      []Annotation   `gorm:"foreignKey:UserID" json:"annotations,omitempty"`
	Collections      []Collection   `gorm:"foreignKey:UserID" json:"collections,omitempty"`
}
