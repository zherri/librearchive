package http

import (
	"errors"
	stdhttp "net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/librearchive/librearchive/internal/models"
	"github.com/librearchive/librearchive/internal/security/passphrase"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (s *Server) createUser(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var input createUserInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(); err != nil {
		fail(w, 400, err.Error())
		return
	}
	user, generatedPassphrase, err := newUser(input, models.RoleReader)
	if err != nil {
		internalError(w, err)
		return
	}
	if err := s.db.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			fail(w, 409, "username is already registered")
			return
		}
		internalError(w, err)
		return
	}
	respond(w, 201, map[string]interface{}{"user": user, "passphrase": generatedPassphrase})
}
func (s *Server) listUsers(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var users []models.User
	paginated(w, r, s.db.Order("created_at desc"), &users)
}

type updateUserInput struct {
	Name     *string      `json:"name"`
	Role     *models.Role `json:"role"`
	IsActive *bool        `json:"isActive"`
}

type updateMyProfileInput struct {
	Name string `json:"name"`
}

func (s *Server) updateMyProfile(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var input updateMyProfileInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if strings.TrimSpace(input.Name) == "" {
		fail(w, 400, "name is required")
		return
	}
	user := currentUser(r)
	user.Name = strings.TrimSpace(input.Name)
	if err := s.db.Save(&user).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, user)
}

func (s *Server) updateUser(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := s.findUser(w, chi.URLParam(r, "userID"))
	if !ok {
		return
	}
	var input updateUserInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.Role != nil && *input.Role != models.RoleAdmin && *input.Role != models.RoleReader {
		fail(w, 400, "role must be admin or reader")
		return
	}
	resultingRole, resultingActive := user.Role, user.IsActive
	if input.Role != nil {
		resultingRole = *input.Role
	}
	if input.IsActive != nil {
		resultingActive = *input.IsActive
	}
	if user.Role == models.RoleAdmin && user.IsActive && (resultingRole != models.RoleAdmin || !resultingActive) && !s.canRemoveActiveAdmin(user) {
		fail(w, 409, "cannot remove the final active administrator")
		return
	}
	if input.Name != nil {
		if strings.TrimSpace(*input.Name) == "" {
			fail(w, 400, "name cannot be empty")
			return
		}
		user.Name = strings.TrimSpace(*input.Name)
	}
	if input.Role != nil {
		user.Role = *input.Role
	}
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}
	if err := s.db.Save(&user).Error; err != nil {
		internalError(w, err)
		return
	}
	if (input.IsActive != nil && !*input.IsActive) || input.Role != nil {
		if err := s.db.Where("user_id = ?", user.ID).Delete(&models.AuthSession{}).Error; err != nil {
			internalError(w, err)
			return
		}
	}
	respond(w, 200, user)
}

func (s *Server) resetUserPassphrase(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := s.findUser(w, chi.URLParam(r, "userID"))
	if !ok {
		return
	}
	generatedPassphrase, err := passphrase.Generate()
	if err != nil {
		internalError(w, err)
		return
	}
	hash, err := bcrypt.GenerateFromPassword(passphrase.CredentialBytes(generatedPassphrase), bcrypt.DefaultCost)
	if err != nil {
		internalError(w, err)
		return
	}
	user.PassphraseHash = string(hash)
	if err := s.db.Save(&user).Error; err != nil {
		internalError(w, err)
		return
	}
	if err := s.db.Where("user_id = ?", user.ID).Delete(&models.AuthSession{}).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, map[string]interface{}{"user": user, "passphrase": generatedPassphrase})
}

func (s *Server) deleteUser(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user, ok := s.findUser(w, chi.URLParam(r, "userID"))
	if !ok {
		return
	}
	if user.Role == models.RoleAdmin && user.IsActive && !s.canRemoveActiveAdmin(user) {
		fail(w, 409, "cannot delete the final active administrator")
		return
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM collection_books WHERE collection_id IN (SELECT id FROM collections WHERE user_id = ?)", user.ID).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.Note{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.Highlight{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.Favorite{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.ReadingProgress{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.AuthSession{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.Collection{}).Error; err != nil {
			return err
		}
		return tx.Delete(&user).Error
	}); err != nil {
		internalError(w, err)
		return
	}
	w.WriteHeader(204)
}

func (s *Server) canRemoveActiveAdmin(user models.User) bool {
	var count int64
	if err := s.db.Model(&models.User{}).Where("role = ? AND is_active = ?", models.RoleAdmin, true).Count(&count).Error; err != nil {
		return false
	}
	return user.Role != models.RoleAdmin || !user.IsActive || count > 1
}
