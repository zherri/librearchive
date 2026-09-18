package http

import (
	"errors"
	stdhttp "net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/librearchive/librearchive/internal/models"
	"gorm.io/gorm"
)

type collectionInput struct {
	Name string `json:"name"`
}

func (input collectionInput) validate() error {
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("name is required")
	}
	if len(strings.TrimSpace(input.Name)) > 120 {
		return errors.New("name must not exceed 120 characters")
	}
	return nil
}
func (s *Server) listCollections(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var collections []models.Collection
	if err := s.db.Where("user_id = ?", currentUser(r).ID).Order("created_at desc").Find(&collections).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, collections)
}
func (s *Server) createCollection(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var input collectionInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(); err != nil {
		fail(w, 400, err.Error())
		return
	}
	collection := models.Collection{UserID: currentUser(r).ID, Name: strings.TrimSpace(input.Name)}
	if err := s.db.Create(&collection).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 201, collection)
}
func (s *Server) getCollection(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	collection, ok := s.findOwnedCollection(w, r)
	if !ok {
		return
	}
	var items []models.CollectionBook
	if err := s.db.Preload("Book").Where("collection_id = ?", collection.ID).Find(&items).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, map[string]interface{}{"collection": collection, "items": items})
}
func (s *Server) updateCollection(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	collection, ok := s.findOwnedCollection(w, r)
	if !ok {
		return
	}
	var input collectionInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(); err != nil {
		fail(w, 400, err.Error())
		return
	}
	collection.Name = strings.TrimSpace(input.Name)
	if err := s.db.Save(&collection).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, collection)
}
func (s *Server) deleteCollection(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	collection, ok := s.findOwnedCollection(w, r)
	if !ok {
		return
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("collection_id = ?", collection.ID).Delete(&models.CollectionBook{}).Error; err != nil {
			return err
		}
		return tx.Delete(&collection).Error
	}); err != nil {
		internalError(w, err)
		return
	}
	w.WriteHeader(204)
}
func (s *Server) addBookToCollection(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	collection, ok := s.findOwnedCollection(w, r)
	if !ok {
		return
	}
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	item := models.CollectionBook{CollectionID: collection.ID, BookID: book.ID}
	if err := s.db.Where("collection_id = ? AND book_id = ?", item.CollectionID, item.BookID).FirstOrCreate(&item).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, item)
}
func (s *Server) removeBookFromCollection(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	collection, ok := s.findOwnedCollection(w, r)
	if !ok {
		return
	}
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	result := s.db.Where("collection_id = ? AND book_id = ?", collection.ID, book.ID).Delete(&models.CollectionBook{})
	if result.Error != nil {
		internalError(w, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		fail(w, 404, "book is not in this collection")
		return
	}
	w.WriteHeader(204)
}
func (s *Server) findOwnedCollection(w stdhttp.ResponseWriter, r *stdhttp.Request) (models.Collection, bool) {
	var collection models.Collection
	if err := s.db.Where("id = ? AND user_id = ?", chi.URLParam(r, "collectionID"), currentUser(r).ID).First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(w, 404, "collection not found")
		} else {
			internalError(w, err)
		}
		return models.Collection{}, false
	}
	return collection, true
}
