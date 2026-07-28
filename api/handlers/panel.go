package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/zherri/librearchive/models"
	"github.com/zherri/librearchive/repositories"
	"gorm.io/gorm"
)

type panelHandler struct {
	ur repositories.IDBRepository[models.User]
	br repositories.IDBRepository[models.Book]
}

func NewPanelHandler(db *gorm.DB) *panelHandler {
	return &panelHandler{
		ur: repositories.NewDBRepository[models.User](db),
		br: repositories.NewDBRepository[models.Book](db),
	}
}

func (h *panelHandler) Overview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	usersCount, err := h.ur.Count(ctx)
	if err != nil {
		http.Error(w, "error loading users count", http.StatusInternalServerError)
		return
	}

	booksCount, err := h.br.Count(ctx)
	if err != nil {
		http.Error(w, "error loading books count", http.StatusInternalServerError)
		return
	}

	users, err := h.ur.GetAll(ctx)
	if err != nil {
		http.Error(w, "error loading users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"users_count": usersCount,
		"books_count": booksCount,
		"users":       users,
	})
}
