package models

import "time"

type Book struct {
	ID               uint      `gorm:"primaryKey" json:"book_id"`
	Title            string    `gorm:"not null;index;type:varchar(100)" json:"title"`
	Authors          string    `gorm:"not null;type:varchar(100)" json:"authors"`
	Genre            string    `gorm:"not null;type:varchar(255)" json:"genre"`
	Publisher        string    `gorm:"not null;type:varchar(100)" json:"publisher"`
	PublicationDate  string    `gorm:"not null;type:varchar(20)" json:"publication_date"`
	FilePath         string    `gorm:"not null;type:varchar(255)" json:"-"`
	CoverPath        string    `gorm:"not null;type:varchar(255)" json:"cover_path"`
	RegistrationDate time.Time `gorm:"type:timestamptz;autoCreateTime" json:"registration_date"`
}
