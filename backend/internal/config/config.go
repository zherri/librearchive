package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ListenAddress  string
	DatabasePath   string
	StoragePath    string
	JWTSecret      string
	AllowedOrigins []string
	MaxUploadBytes int64
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, err
	}
	cfg := Config{
		ListenAddress:  valueOrDefault("LISTEN_ADDRESS", ":8080"),
		DatabasePath:   valueOrDefault("DATABASE_PATH", "./data/librearchive.db"),
		StoragePath:    valueOrDefault("STORAGE_PATH", "./data/storage"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		AllowedOrigins: splitCSV(valueOrDefault("CORS_ALLOWED_ORIGINS", "")),
		MaxUploadBytes: 250 << 20,
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}
	if value := strings.TrimSpace(os.Getenv("MAX_UPLOAD_BYTES")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed < 1 {
			return Config{}, errors.New("MAX_UPLOAD_BYTES must be a positive integer")
		}
		cfg.MaxUploadBytes = parsed
	}
	if err := os.MkdirAll(filepath.Dir(cfg.DatabasePath), 0o750); err != nil {
		return Config{}, err
	}
	if err := os.MkdirAll(filepath.Join(cfg.StoragePath, "books"), 0o750); err != nil {
		return Config{}, err
	}
	if err := os.MkdirAll(filepath.Join(cfg.StoragePath, "covers"), 0o750); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	values := strings.Split(value, ",")
	result := make([]string, 0, len(values))
	for _, item := range values {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func valueOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
