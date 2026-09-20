package http

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	stdhttp "net/http"
	"strconv"
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
	parsed, err := jwt.ParseWithClaims(input.RefreshToken, &refreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.JWTSecret), nil
	})
	if err != nil || !parsed.Valid {
		fail(w, 401, "invalid or expired refresh token")
		return
	}
	tokenClaims, ok := parsed.Claims.(*refreshClaims)
	if !ok || tokenClaims.SessionID == 0 {
		fail(w, 401, "invalid or expired refresh token")
		return
	}
	var session models.AuthSession
	if err := s.db.Where("id = ? AND refresh_token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenClaims.SessionID, hashRefreshToken(input.RefreshToken), time.Now().UTC()).First(&session).Error; err != nil {
		fail(w, 401, "invalid or expired refresh token")
		return
	}
	var user models.User
	if err := s.db.First(&user, session.UserID).Error; err != nil || !user.IsActive || tokenClaims.Subject != strconv.FormatUint(uint64(user.ID), 10) {
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
	expiresAt := time.Now().UTC().Add(sessionLifetime(user.Role))
	session := models.AuthSession{
		UserID:           user.ID,
		RefreshTokenHash: hashRefreshToken(fmt.Sprintf("pending:%d:%d", user.ID, time.Now().UnixNano())),
		ExpiresAt:        expiresAt,
	}
	if err := s.db.Create(&session).Error; err != nil {
		return nil, err
	}
	accessToken, err := s.tokenFor(user, session.ID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.refreshTokenFor(user, session.ID, expiresAt)
	if err != nil {
		return nil, err
	}
	if err := s.db.Model(&session).Update("refresh_token_hash", hashRefreshToken(refreshToken)).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"token": accessToken, "refreshToken": refreshToken, "user": user}, nil
}

func sessionLifetime(role models.Role) time.Duration {
	if role == models.RoleAdmin {
		return 24 * time.Hour
	}
	return 365 * 24 * time.Hour
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
