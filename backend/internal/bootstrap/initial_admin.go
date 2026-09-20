// Package bootstrap initializes data required for a new LibreArchive server.
package bootstrap

import (
	"github.com/librearchive/librearchive/internal/models"
	"github.com/librearchive/librearchive/internal/security/passphrase"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const initialAdministratorUsername = "admin"

type InitialAdministrator struct {
	Created    bool
	Username   string
	Passphrase string
}

// EnsureInitialAdministrator creates the default administrator only when the
// database has no administrator. The passphrase is returned solely at creation
// time so the caller can display it to the server operator.
func EnsureInitialAdministrator(db *gorm.DB) (InitialAdministrator, error) {
	result := InitialAdministrator{Username: initialAdministratorUsername}
	err := db.Transaction(func(tx *gorm.DB) error {
		var administratorCount int64
		if err := tx.Model(&models.User{}).
			Where("role = ?", models.RoleAdmin).
			Count(&administratorCount).Error; err != nil {
			return err
		}
		if administratorCount > 0 {
			return nil
		}

		generatedPassphrase, err := passphrase.Generate()
		if err != nil {
			return err
		}
		hash, err := bcrypt.GenerateFromPassword(
			passphrase.CredentialBytes(generatedPassphrase),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return err
		}
		administrator := models.User{
			Username:       initialAdministratorUsername,
			Name:           initialAdministratorUsername,
			PassphraseHash: string(hash),
			Role:           models.RoleAdmin,
			IsActive:       true,
		}
		if err := tx.Create(&administrator).Error; err != nil {
			return err
		}
		result.Created = true
		result.Passphrase = generatedPassphrase
		return nil
	})
	return result, err
}
