package bootstrap_test

import (
	"path/filepath"
	"testing"

	"github.com/librearchive/librearchive/internal/bootstrap"
	"github.com/librearchive/librearchive/internal/database"
	"github.com/librearchive/librearchive/internal/models"
	"github.com/librearchive/librearchive/internal/security/passphrase"
	"golang.org/x/crypto/bcrypt"
)

func TestEnsureInitialAdministratorCreatesOnlyOnce(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "library.db"))
	if err != nil {
		t.Fatal(err)
	}
	created, err := bootstrap.EnsureInitialAdministrator(db)
	if err != nil {
		t.Fatal(err)
	}
	if !created.Created || created.Username != "admin" || created.Passphrase == "" {
		t.Fatalf("unexpected initial administrator: %#v", created)
	}
	var administrator models.User
	if err := db.Where("username = ?", "admin").First(&administrator).Error; err != nil {
		t.Fatal(err)
	}
	if administrator.Role != models.RoleAdmin || !administrator.IsActive {
		t.Fatalf("unexpected administrator: %#v", administrator)
	}
	if bcrypt.CompareHashAndPassword([]byte(administrator.PassphraseHash), passphrase.CredentialBytes(created.Passphrase)) != nil {
		t.Fatal("generated passphrase does not match its stored hash")
	}

	again, err := bootstrap.EnsureInitialAdministrator(db)
	if err != nil {
		t.Fatal(err)
	}
	if again.Created || again.Passphrase != "" {
		t.Fatalf("administrator was recreated: %#v", again)
	}
}
