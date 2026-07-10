package handlers

import (
	"net/http"

	"gorm.io/gorm"
)

type authHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *authHandler {
	return &authHandler{
		db: db,
	}
}

type registerDTO struct {
	Username string `json:"username"`
}

func (ah *authHandler) Register(w http.ResponseWriter, r *http.Request) {
}

func (ah *authHandler) Login(w http.ResponseWriter, r *http.Request) {
}
