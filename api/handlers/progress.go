package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zherri/librearchive/models"
	"github.com/zherri/librearchive/repositories"
	"gorm.io/gorm"
)

type progressHandler struct {
	pr repositories.IDBRepository[models.Progress]
}

func NewProgressHandler(db *gorm.DB) *progressHandler {
	return &progressHandler{
		pr: repositories.NewDBRepository[models.Progress](db),
	}
}

func (h *progressHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}
	search := strings.TrimSpace(query.Get("search"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	response, err := h.pr.GetPaginatedWithAssoc(ctx, page, limit, search, "Book")
	if err != nil {
		http.Error(w, "error trying to get books", http.StatusInternalServerError)
		log.Printf("Error trying to get books: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *progressHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}

	progress, err := h.pr.FindByIDWithAssoc(r.Context(), uint(id), "Book")
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
		"progress": progress,
	})
}
