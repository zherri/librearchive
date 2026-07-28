package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/zherri/librearchive/infra"
	"github.com/zherri/librearchive/models"
	"github.com/zherri/librearchive/repositories"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type authHandler struct {
	ur repositories.IDBRepository[models.User]
}

func NewAuthHandler(db *gorm.DB) *authHandler {
	return &authHandler{
		ur: repositories.NewDBRepository[models.User](db),
	}
}

type registerDTO struct {
	Username string `json:"username"`
}

func (ah *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var dto registerDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if dto.Username == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	users, err := ah.ur.Find(ctx, "username = ?", dto.Username)
	if err != nil {
		http.Error(w, "error searching existent user", http.StatusInternalServerError)
		log.Printf("Error searching existent user: %v", err)
		return
	}

	if len(users) != 0 {
		http.Error(w, "user already exist", http.StatusConflict)
		return
	}

	passphrase, err := infra.GeneratePassphrase()
	if err != nil {
		http.Error(w, "error generating passphrase", http.StatusInternalServerError)
		log.Printf("Error generating passphrase: %v", err)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(passphrase), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error hashing passphrase", http.StatusInternalServerError)
		log.Printf("Error hashing passphrase: %v", err)
		return
	}

	newUser := models.User{
		Username:   dto.Username,
		Passphrase: string(hashed),
	}

	if err := ah.ur.Create(ctx, &newUser); err != nil {
		http.Error(w, "error creating new user", http.StatusInternalServerError)
		log.Printf("Error creating new user: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(loginDTO{
		Username:   newUser.Username,
		Passphrase: passphrase,
	})
}

type loginDTO struct {
	Username   string `json:"username"`
	Passphrase string `json:"passphrase"`
}

func (ah *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	var dto loginDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if dto.Username == "" || dto.Passphrase == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	users, err := ah.ur.Find(r.Context(), "username = ?", dto.Username)
	if err != nil {
		http.Error(w, "error finding user", http.StatusInternalServerError)
		log.Printf("Error finding user: %v", err)
		return
	}

	if len(users) == 0 {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(users[0].Passphrase), []byte(dto.Passphrase)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := infra.CreateAccessToken(users[0].ID, users[0].IsAdmin)
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		log.Printf("Error generating token: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"token": token,
		"user":  users[0],
	})
}
