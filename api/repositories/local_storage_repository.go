package repositories

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ILocalStorageRepository interface {
	AddBook(bookFilename string, bookFile io.Reader, coverFilename string, coverFile io.Reader) (string, string, error)
	RemoveBook(bookFilename string, coverFilename string) error
}

type localStorageRepository struct {
	booksDir  string
	coversDir string
}

func NewLocalStorageRepository() (ILocalStorageRepository, error) {
	dataPath := os.Getenv("STORAGE_PATH")

	booksDir := filepath.Join(dataPath, "books")
	if err := os.MkdirAll(booksDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	coversDir := filepath.Join(dataPath, "covers")
	if err := os.MkdirAll(coversDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &localStorageRepository{
		booksDir:  booksDir,
		coversDir: coversDir,
	}, nil
}

func (r *localStorageRepository) AddBook(bookFilename string, bookData io.Reader, coverFilename string, coverData io.Reader) (string, string, error) {
	rb := make([]byte, 16)
	_, _ = rand.Read(rb)
	randHex := hex.EncodeToString(rb)
	newBookFilename := randHex + strings.ToLower(filepath.Ext(bookFilename))
	newCoverFilename := randHex + "_cover" + strings.ToLower(filepath.Ext(coverFilename))

	bookPath := filepath.Join(r.booksDir, newBookFilename)
	coverPath := filepath.Join(r.coversDir, newCoverFilename)

	bookFile, err := os.Create(bookPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file on disk: %w", err)
	}
	defer bookFile.Close()

	if _, err := io.Copy(bookFile, bookData); err != nil {
		return "", "", fmt.Errorf("failed to save file contents: %w", err)
	}

	coverFile, err := os.Create(coverPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file on disk: %w", err)
	}
	defer coverFile.Close()

	if _, err := io.Copy(coverFile, coverData); err != nil {
		return "", "", fmt.Errorf("failed to save file contents: %w", err)
	}

	return newBookFilename, newCoverFilename, nil
}

func (r *localStorageRepository) RemoveBook(bookFilename string, coverFilename string) error {
	bookPath := filepath.Join(r.booksDir, bookFilename)
	coverPath := filepath.Join(r.coversDir, coverFilename)

	if err := os.Remove(bookPath); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	if err := os.Remove(coverPath); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	return nil
}
