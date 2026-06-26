package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zherri/librearchive/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
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
		log.Fatalf("Falha crítica ao conectar ao banco de dados: %v", err)
	}

	err = db.AutoMigrate(
		&models.Annotation{},
		&models.BookProgress{},
		&models.Collection{},
		&models.CollectionBook{},
		&models.Book{},
		&models.User{},
	)
	if err != nil {
		log.Fatalf("Falha crítica ao executar o AutoMigrate do GORM: %v", err)
	}

	return db
}
