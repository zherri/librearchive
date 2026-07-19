package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/zherri/librearchive/models"
	"github.com/zherri/librearchive/repositories"
	"gorm.io/gorm"
)

type bookHandler struct {
	br repositories.IDBRepository[models.Book]
}

func NewBookHandler(db *gorm.DB) *bookHandler {
	return &bookHandler{
		br: repositories.NewDBRepository[models.Book](db),
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
}

type UpdateBookDTO struct {
	Title     *string `json:"title"`
	Authors   *string `json:"authors"`
	Genre     *string `json:"genre"`
	Publisher *string `json:"publisher"`
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
	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}
}
