package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/librearchive/librearchive/internal/models"
)

func Open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.New(log.New(os.Stderr, "", log.LstdFlags), logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		}),
	})
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
	if hasColumn(db, "books", "author") {
		if err := db.Exec("UPDATE books SET authors = author WHERE authors = ''").Error; err != nil {
			return nil, err
		}
	}
	return db, nil
}

func hasColumn(db *gorm.DB, table, name string) bool {
	var columns []struct{ Name string }
	if err := db.Raw("PRAGMA table_info(" + table + ")").Scan(&columns).Error; err != nil {
		return false
	}
	for _, column := range columns {
		if column.Name == name {
			return true
		}
	}
	return false
}
