package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/librearchive/librearchive/internal/models"
)

func Open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&models.User{}, &models.Book{}, &models.ReadingProgress{}, &models.Favorite{},
		&models.Collection{}, &models.CollectionBook{}, &models.Highlight{}, &models.Note{},
		&models.Category{}, &models.BookCategory{}, &models.Tag{}, &models.BookTag{},
		&models.AuthSession{},
	); err != nil {
		return nil, err
	}
	return db, nil
}
