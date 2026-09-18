package models

import "time"

type ReadingStatus string

const (
	ReadingStatusUnread     ReadingStatus = "unread"
	ReadingStatusInProgress ReadingStatus = "in_progress"
	ReadingStatusFinished   ReadingStatus = "finished"
)

type ReadingProgress struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	UserID          uint          `gorm:"not null;uniqueIndex:idx_user_book_progress" json:"userId"`
	BookID          uint          `gorm:"not null;uniqueIndex:idx_user_book_progress" json:"bookId"`
	Book            Book          `gorm:"foreignKey:BookID" json:"book,omitempty"`
	Status          ReadingStatus `gorm:"size:20;not null;default:unread;index" json:"status"`
	CurrentPage     int           `gorm:"not null;default:0" json:"currentPage"`
	ProgressPercent float64       `gorm:"not null;default:0" json:"progressPercent"`
	IsDownloaded    bool          `gorm:"not null;default:false;index" json:"isDownloaded"`
	LastReadAt      *time.Time    `json:"lastReadAt,omitempty"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}
