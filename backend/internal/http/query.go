package http

import (
	"errors"
	stdhttp "net/http"
	"strconv"
	"strings"

	"github.com/librearchive/librearchive/internal/models"
	"gorm.io/gorm"
)

func pagination(r *stdhttp.Request) (int, int) {
	page, size := 1, 20
	if value, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && value > 0 {
		page = value
	}
	if value, err := strconv.Atoi(r.URL.Query().Get("pageSize")); err == nil && value > 0 && value <= 100 {
		size = value
	}
	return page, size
}
func optionalPositiveInt(value string) (*int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return nil, errors.New("not positive")
	}
	return &parsed, nil
}
func (s *Server) findUser(w stdhttp.ResponseWriter, id string) (models.User, bool) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(w, 404, "user not found")
		} else {
			internalError(w, err)
		}
		return models.User{}, false
	}
	return user, true
}
func (s *Server) findBook(w stdhttp.ResponseWriter, id string) (models.Book, bool) {
	var book models.Book
	if err := s.db.First(&book, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(w, 404, "book not found")
		} else {
			internalError(w, err)
		}
		return models.Book{}, false
	}
	return book, true
}
