package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zherri/librearchive/models"
	"github.com/zherri/librearchive/repositories"
	"gorm.io/gorm"
)

type bookHandler struct {
	br  repositories.IDBRepository[models.Book]
	lsr repositories.ILocalStorageRepository
}

func NewBookHandler(db *gorm.DB, lsr repositories.ILocalStorageRepository) *bookHandler {
	return &bookHandler{
		br:  repositories.NewDBRepository[models.Book](db),
		lsr: lsr,
	}
}

func (bh *bookHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	var books []models.Book
	var err error

	if _, ok := query["search"]; !ok {
		books, err = bh.br.GetAll(ctx)
		if err != nil {
			http.Error(w, "error trying to get books", http.StatusInternalServerError)
			log.Printf("Error trying to get books: %v", err)
			return
		}
	} else {
		search := query.Get("search")
		if search == "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		search = "%" + search + "%"

		sqlQuery := "title LIKE ? OR authors LIKE ? OR genre LIKE ? OR publisher LIKE ?"

		books, err = bh.br.Find(ctx, sqlQuery, search, search, search, search)
		if err != nil {
			http.Error(w, "error trying to get books", http.StatusInternalServerError)
			log.Printf("Error trying to get books: %v", err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"books": books,
	})
}

func (bh *bookHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}

	book, err := bh.br.FindByID(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "book not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error trying to get book", http.StatusInternalServerError)
		log.Printf("Error trying to get book: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"book": book,
	})
}

func (bh *bookHandler) Upload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	bookFile, bookFileHeader, err := r.FormFile("book")
	if err != nil {
		http.Error(w, "invalid book file", http.StatusBadRequest)
		return
	}
	defer bookFile.Close()

	bookExt := strings.ToLower(filepath.Ext(bookFileHeader.Filename))
	if bookExt != ".epub" && bookExt != ".pdf" {
		http.Error(w, "only epub and pdf are supported", http.StatusBadRequest)
		return
	}

	coverFile, coverFileHeader, err := r.FormFile("cover")
	if err != nil {
		http.Error(w, "invalid cover file", http.StatusBadRequest)
		return
	}
	defer coverFile.Close()

	coverExt := strings.ToLower(filepath.Ext(coverFileHeader.Filename))
	if coverExt != ".png" && coverExt != ".jpg" && coverExt != ".jpeg" {
		http.Error(w, "only png, jpg and jpeg are supported", http.StatusBadRequest)
		return
	}

	bookFilename, coverFilename, err := bh.lsr.AddBook(bookFileHeader.Filename, bookFile, coverFileHeader.Filename, coverFile)
	if err != nil {
		http.Error(w, "failed to save book", http.StatusInternalServerError)
		log.Printf("Failed to save book: %v", err)
		return
	}

	title := r.FormValue("title")
	authors := r.FormValue("authors")
	genre := r.FormValue("genre")
	publisher := r.FormValue("publisher")
	publicationDate := r.FormValue("publication_date")

	book := models.Book{
		Title:           title,
		Authors:         authors,
		Genre:           genre,
		Publisher:       publisher,
		PublicationDate: publicationDate,
		Filename:        bookFilename,
		CoverName:       coverFilename,
	}

	if err := bh.br.Create(r.Context(), &book); err != nil {
		http.Error(w, "error creating new book", http.StatusInternalServerError)
		log.Printf("Error creating new book: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type UpdateBookDTO struct {
	Title           *string `json:"title"`
	Authors         *string `json:"authors"`
	Genre           *string `json:"genre"`
	Publisher       *string `json:"publisher"`
	PublicationDate *string `json:"publication_date"`
}

func (bh *bookHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		http.Error(w, "invalid url parameters", http.StatusBadRequest)
		return
	}

	var dto UpdateBookDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	book, err := bh.br.FindByID(ctx, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "book not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error trying to get book", http.StatusInternalServerError)
		log.Printf("Error trying to get book: %v", err)
		return
	}

	if dto.Title != nil && *dto.Title != "" {
		book.Title = *dto.Title
	}
	if dto.Authors != nil && *dto.Authors != "" {
		book.Authors = *dto.Authors
	}
	if dto.Genre != nil && *dto.Genre != "" {
		book.Genre = *dto.Genre
	}
	if dto.Publisher != nil && *dto.Publisher != "" {
		book.Publisher = *dto.Publisher
	}
	if dto.PublicationDate != nil && *dto.PublicationDate != "" {
		book.PublicationDate = *dto.PublicationDate
	}

	if err := bh.br.Save(ctx, book); err != nil {
		http.Error(w, "error trying to save book", http.StatusInternalServerError)
		log.Printf("Error trying to save book: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"book": book,
	})
}

func (bh *bookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}

	book, err := bh.br.FindByID(ctx, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "book not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error trying to get book", http.StatusInternalServerError)
		log.Printf("Error trying to get book: %v", err)
		return
	}

	if err := bh.lsr.RemoveBook(book.Filename, book.CoverName); err != nil {
		http.Error(w, "failed to remove book", http.StatusInternalServerError)
		log.Printf("Failed to remove book: %v", err)
		return
	}

	if err := bh.br.Delete(ctx, uint(id)); err != nil {
		http.Error(w, "error trying to delete book", http.StatusInternalServerError)
		log.Printf("Error trying to delete book: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
