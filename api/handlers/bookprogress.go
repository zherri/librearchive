package handlers

import (
	"net/http"

	"github.com/zherri/librearchive/models"
	"github.com/zherri/librearchive/repositories"
	"gorm.io/gorm"
)

type bookProgressHandler struct {
	bpr repositories.IDBRepository[models.BookProgress]
}

func NewBookProgressHandler(db *gorm.DB) *bookProgressHandler {
	return &bookProgressHandler{
		bpr: repositories.NewDBRepository[models.BookProgress](db),
	}
}

func (bph *bookProgressHandler) Get(w http.ResponseWriter, r *http.Request) {
}
