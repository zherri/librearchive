package http

import (
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"

	"github.com/librearchive/librearchive/internal/models"
)

func (s *Server) listFavorites(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user := currentUser(r)
	var favorites []models.Favorite
	if err := s.db.Preload("Book").Where("user_id = ?", user.ID).Order("created_at desc").Find(&favorites).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, favorites)
}
func (s *Server) addFavorite(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	favorite := models.Favorite{UserID: currentUser(r).ID, BookID: book.ID}
	if err := s.db.Where("user_id = ? AND book_id = ?", favorite.UserID, favorite.BookID).FirstOrCreate(&favorite).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, favorite)
}
func (s *Server) removeFavorite(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	result := s.db.Where("user_id = ? AND book_id = ?", currentUser(r).ID, book.ID).Delete(&models.Favorite{})
	if result.Error != nil {
		internalError(w, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		fail(w, 404, "favorite not found")
		return
	}
	w.WriteHeader(204)
}
