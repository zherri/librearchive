package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/librearchive/librearchive/internal/models"
)

type claims struct {
	Role      models.Role `json:"role"`
	SessionID uint        `json:"sid"`
	jwt.RegisteredClaims
}

type refreshClaims struct {
	SessionID uint `json:"sid"`
	jwt.RegisteredClaims
}
type userContextKey struct{}

func (s *Server) authenticate(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		value := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if value == r.Header.Get("Authorization") || value == "" {
			fail(w, 401, "missing bearer token")
			return
		}
		parsed, err := jwt.ParseWithClaims(value, &claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(s.config.JWTSecret), nil
		})
		if err != nil || !parsed.Valid {
			fail(w, 401, "invalid or expired token")
			return
		}
		tokenClaims := parsed.Claims.(*claims)
		if tokenClaims.SessionID == 0 {
			fail(w, 401, "invalid token session")
			return
		}
		var session models.AuthSession
		if err := s.db.Where("id = ? AND revoked_at IS NULL AND expires_at > ?", tokenClaims.SessionID, time.Now().UTC()).First(&session).Error; err != nil {
			fail(w, 401, "revoked or expired session")
			return
		}
		var user models.User
		if err := s.db.First(&user, tokenClaims.Subject).Error; err != nil || !user.IsActive {
			fail(w, 401, "invalid or inactive user")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey{}, user)))
	})
}
func requireRole(role models.Role) func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			if currentUser(r).Role != role {
				fail(w, 403, "administrator access is required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
func currentUser(r *stdhttp.Request) models.User {
	user, _ := r.Context().Value(userContextKey{}).(models.User)
	return user
}
func (s *Server) tokenFor(user models.User, sessionID uint) (string, error) {
	now := time.Now()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims{Role: user.Role, SessionID: sessionID, RegisteredClaims: jwt.RegisteredClaims{Subject: strconv.FormatUint(uint64(user.ID), 10), ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), IssuedAt: jwt.NewNumericDate(now)}}).SignedString([]byte(s.config.JWTSecret))
}

func (s *Server) refreshTokenFor(user models.User, sessionID uint, expiresAt time.Time) (string, error) {
	now := time.Now()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims{SessionID: sessionID, RegisteredClaims: jwt.RegisteredClaims{Subject: strconv.FormatUint(uint64(user.ID), 10), ExpiresAt: jwt.NewNumericDate(expiresAt), IssuedAt: jwt.NewNumericDate(now)}}).SignedString([]byte(s.config.JWTSecret))
}
