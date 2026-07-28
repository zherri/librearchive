package infra

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zherri/librearchive/models"
	"golang.org/x/crypto/bcrypt"
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
		&models.Progress{},
		&models.Annotation{},
	)
	if err != nil {
		log.Fatalf("Error migrating entities: %v", err)
	}

	return db
}

func CreateAdminUser(db *gorm.DB) {
	adminName := os.Getenv("ADMIN_NAME")
	adminPassphrase := os.Getenv("ADMIN_PASSPHRASE")

	if adminName == "" || adminPassphrase == "" {
		log.Fatalf("ALERT: ADMIN_NAME and ADMIN_PASSPHRASE not defined in .env. Please define to be able to do admin things.")
	}

	var exists bool
	err := db.Model(&models.User{}).Select("count(1) > 0").Where("username = ?", adminName).Find(&exists).Error
	if err != nil {
		log.Fatalf("Error checking admin existence: %v", err)
	}

	if exists {
		return
	}

	log.Printf("Admin '%s' not found in database. Automatically creating account...", adminName)

	hashed, err := bcrypt.GenerateFromPassword([]byte(adminPassphrase), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Error generating admin passphrase hash: %v", err)
	}

	admin := models.User{
		Username:   adminName,
		Passphrase: string(hashed),
		IsAdmin:    true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("Error saving admin user: %v", err)
	}

	log.Println("Admin created successfully!")
}
