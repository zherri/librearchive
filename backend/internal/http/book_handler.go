package http

import (
	"bytes"
	"fmt"
	"io"
	stdhttp "net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"gorm.io/gorm"

	"github.com/librearchive/librearchive/internal/models"
)

func (s *Server) listBooks(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	query := s.db.Model(&models.Book{})
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		term := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(authors) LIKE ?", term, term)
	}
	if categoryID := strings.TrimSpace(r.URL.Query().Get("categoryId")); categoryID != "" {
		query = query.Joins("JOIN book_categories ON book_categories.book_id = books.id").Where("book_categories.category_id = ?", categoryID)
	}
	if tagID := strings.TrimSpace(r.URL.Query().Get("tagId")); tagID != "" {
		query = query.Joins("JOIN book_tags ON book_tags.book_id = books.id").Where("book_tags.tag_id = ?", tagID)
	}
	offset, limit := pagination(r)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		internalError(w, err)
		return
	}
	var books []models.Book
	if err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&books).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, map[string]interface{}{"items": books, "offset": offset, "limit": limit, "total": total})
}
func (s *Server) getBook(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if ok {
		respond(w, 200, book)
	}
}
func (s *Server) createBook(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	r.Body = stdhttp.MaxBytesReader(w, r.Body, s.config.MaxUploadBytes)
	if err := r.ParseMultipartForm(s.config.MaxUploadBytes); err != nil {
		fail(w, 400, "invalid multipart upload or file exceeds the configured limit")
		return
	}
	title, authors := strings.TrimSpace(r.FormValue("title")), strings.TrimSpace(r.FormValue("authors"))
	if title == "" || authors == "" {
		fail(w, 400, "title and authors are required")
		return
	}
	pdf, header, err := r.FormFile("file")
	if err != nil {
		fail(w, 400, "a PDF file is required")
		return
	}
	defer pdf.Close()
	if !strings.EqualFold(filepath.Ext(header.Filename), ".pdf") {
		fail(w, 400, "book file must have a .pdf extension")
		return
	}
	prefix := make([]byte, 512)
	read, err := io.ReadFull(pdf, prefix)
	if err != nil && err != io.ErrUnexpectedEOF {
		internalError(w, err)
		return
	}
	prefix = prefix[:read]
	if !bytes.HasPrefix(prefix, []byte("%PDF-")) || stdhttp.DetectContentType(prefix) != "application/pdf" {
		fail(w, 400, "book file must contain a valid PDF signature")
		return
	}
	stored, err := randomFilename(".pdf")
	if err != nil {
		internalError(w, err)
		return
	}
	bookPath := filepath.Join(s.config.StoragePath, "books", stored)
	if err := copyFile(bookPath, io.MultiReader(bytes.NewReader(prefix), pdf)); err != nil {
		internalError(w, err)
		return
	}
	pageCount, err := pdfapi.PageCountFile(bookPath)
	if err != nil || pageCount < 1 {
		_ = os.Remove(bookPath)
		fail(w, 400, "book file must be a readable PDF")
		return
	}
	book := models.Book{Title: title, Authors: authors, Description: strings.TrimSpace(r.FormValue("description")), Publisher: strings.TrimSpace(r.FormValue("publisher")), StoredFilename: stored, OriginalName: header.Filename, PageCount: &pageCount}
	if year, err := optionalPositiveInt(r.FormValue("publishedYear")); err != nil {
		_ = os.Remove(bookPath)
		fail(w, 400, "publishedYear must be a positive integer")
		return
	} else {
		book.PublishedYear = year
	}
	if cover, coverHeader, err := r.FormFile("cover"); err == nil {
		defer cover.Close()
		extension := strings.ToLower(filepath.Ext(coverHeader.Filename))
		if extension != ".jpg" && extension != ".jpeg" && extension != ".png" && extension != ".webp" {
			_ = os.Remove(bookPath)
			fail(w, 400, "cover must be JPG, PNG, or WEBP")
			return
		}
		coverPrefix := make([]byte, 512)
		coverRead, readErr := io.ReadFull(cover, coverPrefix)
		if readErr != nil && readErr != io.ErrUnexpectedEOF {
			_ = os.Remove(bookPath)
			internalError(w, readErr)
			return
		}
		coverPrefix = coverPrefix[:coverRead]
		mediaType := stdhttp.DetectContentType(coverPrefix)
		if !isAllowedCoverMediaType(mediaType) {
			_ = os.Remove(bookPath)
			fail(w, 400, "cover content must be an image")
			return
		}
		coverName, err := randomFilename(extension)
		if err != nil {
			_ = os.Remove(bookPath)
			internalError(w, err)
			return
		}
		if err := copyFile(filepath.Join(s.config.StoragePath, "covers", coverName), io.MultiReader(bytes.NewReader(coverPrefix), cover)); err != nil {
			_ = os.Remove(bookPath)
			internalError(w, err)
			return
		}
		book.CoverFilename = coverName
	}
	if err := s.db.Create(&book).Error; err != nil {
		_ = os.Remove(bookPath)
		if book.CoverFilename != "" {
			_ = os.Remove(filepath.Join(s.config.StoragePath, "covers", book.CoverFilename))
		}
		internalError(w, err)
		return
	}
	respond(w, 201, book)
}

func isAllowedCoverMediaType(mediaType string) bool {
	return mediaType == "image/jpeg" || mediaType == "image/png" || mediaType == "image/webp"
}

type updateBookInput struct {
	Title         *string `json:"title"`
	Authors       *string `json:"authors"`
	Description   *string `json:"description"`
	Publisher     *string `json:"publisher"`
	PublishedYear *int    `json:"publishedYear"`
}

func (s *Server) updateBook(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	var input updateBookInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.Title != nil {
		if strings.TrimSpace(*input.Title) == "" {
			fail(w, 400, "title cannot be empty")
			return
		}
		book.Title = strings.TrimSpace(*input.Title)
	}
	if input.Authors != nil {
		if strings.TrimSpace(*input.Authors) == "" {
			fail(w, 400, "authors cannot be empty")
			return
		}
		book.Authors = strings.TrimSpace(*input.Authors)
	}
	if input.Description != nil {
		book.Description = strings.TrimSpace(*input.Description)
	}
	if input.Publisher != nil {
		book.Publisher = strings.TrimSpace(*input.Publisher)
	}
	if input.PublishedYear != nil {
		if *input.PublishedYear < 1 {
			fail(w, 400, "publishedYear must be positive")
			return
		}
		book.PublishedYear = input.PublishedYear
	}
	if err := s.db.Save(&book).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, book)
}
func (s *Server) deleteBook(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("book_id = ?", book.ID).Delete(&models.Note{}).Error; err != nil {
			return err
		}
		if err := tx.Where("book_id = ?", book.ID).Delete(&models.Highlight{}).Error; err != nil {
			return err
		}
		if err := tx.Where("book_id = ?", book.ID).Delete(&models.CollectionBook{}).Error; err != nil {
			return err
		}
		if err := tx.Where("book_id = ?", book.ID).Delete(&models.BookCategory{}).Error; err != nil {
			return err
		}
		if err := tx.Where("book_id = ?", book.ID).Delete(&models.BookTag{}).Error; err != nil {
			return err
		}
		if err := tx.Where("book_id = ?", book.ID).Delete(&models.Favorite{}).Error; err != nil {
			return err
		}
		if err := tx.Where("book_id = ?", book.ID).Delete(&models.ReadingProgress{}).Error; err != nil {
			return err
		}
		return tx.Delete(&book).Error
	}); err != nil {
		internalError(w, err)
		return
	}
	_ = os.Remove(filepath.Join(s.config.StoragePath, "books", book.StoredFilename))
	if book.CoverFilename != "" {
		_ = os.Remove(filepath.Join(s.config.StoragePath, "covers", book.CoverFilename))
	}
	w.WriteHeader(204)
}
func (s *Server) serveBook(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	path := filepath.Join(s.config.StoragePath, "books", book.StoredFilename)
	if _, err := os.Stat(path); err != nil {
		fail(w, 404, "book file not found")
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", book.OriginalName))
	stdhttp.ServeFile(w, r, path)
}
func (s *Server) serveCover(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	if book.CoverFilename == "" {
		fail(w, 404, "cover not found")
		return
	}
	stdhttp.ServeFile(w, r, filepath.Join(s.config.StoragePath, "covers", book.CoverFilename))
}
