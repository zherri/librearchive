package infra

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("could not validate credentials")

type CustomClaims struct {
	UserID  uint `json:"user_id"`
	IsAdmin bool `json:"is_admin"`
	jwt.RegisteredClaims
}

func getJWTConfig() ([]byte, jwt.SigningMethod) {
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	return secretKey, jwt.SigningMethodHS512
}

func CreateAccessToken(userID uint, isAdmin bool) (string, error) {
	secretKey, signingMethod := getJWTConfig()

	var exp time.Duration
	if isAdmin {
		exp = 24 * time.Hour
	} else {
		exp = 12 * 30 * 24 * time.Hour
	}

	claims := CustomClaims{
		UserID:  userID,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	return token.SignedString(secretKey)
}

func DecodeToken(tokenString string) (*CustomClaims, error) {
	secretKey, _ := getJWTConfig()
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return secretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
