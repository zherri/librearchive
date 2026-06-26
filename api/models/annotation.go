package models

import (
	"time"
)

type Annotation struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	BookID      string    `gorm:"not null" json:"book_id"`
	CFI         string    `gorm:"type:varchar(255)" json:"cfi,omitempty"`
	PageNumber  int       `json:"page_number,omitempty"`
	ChapterName string    `gorm:"type:varchar(255)" json:"chapter_name"`
	Type        string    `gorm:"type:varchar(20);default:'highlight'" json:"type"`
	TextContent string    `gorm:"type:text;not null" json:"text_content"`
	Note        string    `gorm:"type:text" json:"note,omitempty"`
	Color       string    `gorm:"type:varchar(7);default:'#FFFF00'" json:"color"`
	CreatedAt   time.Time `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
}
