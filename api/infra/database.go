package infra

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zherri/librearchive/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectAndMigrate() *gorm.DB {
	time.Local = time.UTC

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Book{},
		&models.Collection{},
		&models.CollectionBook{},
		&models.BookProgress{},
		&models.Annotation{},
	)
	if err != nil {
		log.Fatalf("Error migrating entities: %v", err)
	}

	return db
}
