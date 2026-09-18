package http

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/librearchive/librearchive/internal/models"
	"github.com/librearchive/librearchive/internal/security/passphrase"
)

type loginInput struct {
	Username   string `json:"username"`
	Passphrase string `json:"passphrase"`
}

func (s *Server) bootstrap(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var count int64
	if err := s.db.Model(&models.User{}).Count(&count).Error; err != nil {
		internalError(w, err)
		return
	}
	if count != 0 {
		fail(w, 409, "bootstrap is only available before the first user is created")
		return
	}
	var input createUserInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(); err != nil {
		fail(w, 400, err.Error())
		return
	}
	user, generatedPassphrase, err := newUser(input, models.RoleAdmin)
	if err != nil {
		internalError(w, err)
		return
	}
	if err := s.db.Create(&user).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 201, map[string]interface{}{"user": user, "passphrase": generatedPassphrase})
}
func (s *Server) login(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var input loginInput
	if !decodeJSON(w, r, &input) {
		return
	}
	var user models.User
	if err := s.db.Where("username = ?", normalizeUsername(input.Username)).First(&user).Error; err != nil || !user.IsActive || bcrypt.CompareHashAndPassword([]byte(user.PassphraseHash), passphrase.CredentialBytes(input.Passphrase)) != nil {
		fail(w, 401, "invalid username or passphrase")
		return
	}
	response, err := s.issueSession(user)
	if err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, response)
}

type refreshInput struct {
	RefreshToken string `json:"refreshToken"`
}

func (s *Server) refresh(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var input refreshInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if strings.TrimSpace(input.RefreshToken) == "" {
		fail(w, 400, "refreshToken is required")
		return
	}
	var session models.AuthSession
	if err := s.db.Where("refresh_token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hashRefreshToken(input.RefreshToken), time.Now().UTC()).First(&session).Error; err != nil {
		fail(w, 401, "invalid or expired refresh token")
		return
	}
	var user models.User
	if err := s.db.First(&user, session.UserID).Error; err != nil || !user.IsActive {
		fail(w, 401, "invalid or inactive user")
		return
	}
	now := time.Now().UTC()
	if err := s.db.Model(&session).Update("revoked_at", &now).Error; err != nil {
		internalError(w, err)
		return
	}
	response, err := s.issueSession(user)
	if err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, response)
}

func (s *Server) logout(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	value := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	parsed, _ := jwt.ParseWithClaims(value, &claims{}, func(token *jwt.Token) (interface{}, error) { return []byte(s.config.JWTSecret), nil })
	if parsed == nil {
		return
	}
	tokenClaims, ok := parsed.Claims.(*claims)
	if !ok {
		return
	}
	now := time.Now().UTC()
	if err := s.db.Model(&models.AuthSession{}).Where("id = ?", tokenClaims.SessionID).Update("revoked_at", &now).Error; err != nil {
		internalError(w, err)
		return
	}
	w.WriteHeader(204)
}

func (s *Server) issueSession(user models.User) (map[string]interface{}, error) {
	refreshToken, err := newRefreshToken()
	if err != nil {
		return nil, err
	}
	session := models.AuthSession{UserID: user.ID, RefreshTokenHash: hashRefreshToken(refreshToken), ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour)}
	if err := s.db.Create(&session).Error; err != nil {
		return nil, err
	}
	accessToken, err := s.tokenFor(user, session.ID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"token": accessToken, "refreshToken": refreshToken, "user": user}, nil
}
func newRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func (s *Server) me(w stdhttp.ResponseWriter, r *stdhttp.Request) { respond(w, 200, currentUser(r)) }

type createUserInput struct {
	Username string      `json:"username"`
	Name     string      `json:"name"`
	Role     models.Role `json:"role"`
}

func (input createUserInput) validate() error {
	username := normalizeUsername(input.Username)
	if len(username) < 3 || len(username) > 80 {
		return errors.New("username must contain between 3 and 80 characters")
	}
	for _, character := range username {
		if !(character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '_' || character == '-') {
			return errors.New("username may contain only letters, numbers, underscores, and hyphens")
		}
	}
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("name is required")
	}
	if input.Role != "" && input.Role != models.RoleAdmin && input.Role != models.RoleReader {
		return errors.New("role must be admin or reader")
	}
	return nil
}
func newUser(input createUserInput, fallback models.Role) (models.User, string, error) {
	generatedPassphrase, err := passphrase.Generate()
	if err != nil {
		return models.User{}, "", err
	}
	hash, err := bcrypt.GenerateFromPassword(passphrase.CredentialBytes(generatedPassphrase), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, "", err
	}
	role := input.Role
	if role == "" {
		role = fallback
	}
	return models.User{Username: normalizeUsername(input.Username), Name: strings.TrimSpace(input.Name), PassphraseHash: string(hash), Role: role, IsActive: true}, generatedPassphrase, nil
}

func normalizeUsername(username string) string { return strings.ToLower(strings.TrimSpace(username)) }
