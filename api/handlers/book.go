package handlers

import (
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
