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
	"gorm.io/gorm"

	"github.com/librearchive/librearchive/internal/models"
)

func (s *Server) listBooks(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	query := s.db.Model(&models.Book{})
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		term := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(author) LIKE ?", term, term)
	}
	if categoryID := strings.TrimSpace(r.URL.Query().Get("categoryId")); categoryID != "" {
		query = query.Joins("JOIN book_categories ON book_categories.book_id = books.id").Where("book_categories.category_id = ?", categoryID)
	}
	if tagID := strings.TrimSpace(r.URL.Query().Get("tagId")); tagID != "" {
		query = query.Joins("JOIN book_tags ON book_tags.book_id = books.id").Where("book_tags.tag_id = ?", tagID)
	}
	page, size := pagination(r)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		internalError(w, err)
		return
	}
	var books []models.Book
	if err := query.Order("created_at desc").Offset((page - 1) * size).Limit(size).Find(&books).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, map[string]interface{}{"items": books, "page": page, "pageSize": size, "total": total})
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
	title, author := strings.TrimSpace(r.FormValue("title")), strings.TrimSpace(r.FormValue("author"))
	if title == "" || author == "" {
		fail(w, 400, "title and author are required")
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
	book := models.Book{Title: title, Author: author, Description: strings.TrimSpace(r.FormValue("description")), Language: strings.TrimSpace(r.FormValue("language")), StoredFilename: stored, OriginalName: header.Filename, UploadedByID: currentUser(r).ID}
	if year, err := optionalPositiveInt(r.FormValue("publishedYear")); err != nil {
		_ = os.Remove(bookPath)
		fail(w, 400, "publishedYear must be a positive integer")
		return
	} else {
		book.PublishedYear = year
	}
	if pages, err := optionalPositiveInt(r.FormValue("pageCount")); err != nil {
		_ = os.Remove(bookPath)
		fail(w, 400, "pageCount must be a positive integer")
		return
	} else {
		book.PageCount = pages
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
	Author        *string `json:"author"`
	Description   *string `json:"description"`
	Language      *string `json:"language"`
	PublishedYear *int    `json:"publishedYear"`
	PageCount     *int    `json:"pageCount"`
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
	if input.Author != nil {
		if strings.TrimSpace(*input.Author) == "" {
			fail(w, 400, "author cannot be empty")
			return
		}
		book.Author = strings.TrimSpace(*input.Author)
	}
	if input.Description != nil {
		book.Description = strings.TrimSpace(*input.Description)
	}
	if input.Language != nil {
		book.Language = strings.TrimSpace(*input.Language)
	}
	if input.PublishedYear != nil {
		if *input.PublishedYear < 1 {
			fail(w, 400, "publishedYear must be positive")
			return
		}
		book.PublishedYear = input.PublishedYear
	}
	if input.PageCount != nil {
		if *input.PageCount < 1 {
			fail(w, 400, "pageCount must be positive")
			return
		}
		book.PageCount = input.PageCount
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
