package http

import (
	"errors"
	stdhttp "net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/librearchive/librearchive/internal/models"
)

type catalogNameInput struct {
	Name string `json:"name"`
}

func (input catalogNameInput) validate(maxLength int) error {
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("name is required")
	}
	if len(strings.TrimSpace(input.Name)) > maxLength {
		return errors.New("name is too long")
	}
	return nil
}

func (s *Server) listCategories(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	var categories []models.Category
	if err := s.db.Order("name").Find(&categories).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, categories)
}
func (s *Server) createCategory(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var input catalogNameInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(120); err != nil {
		fail(w, 400, err.Error())
		return
	}
	category := models.Category{Name: strings.TrimSpace(input.Name)}
	if err := s.db.Create(&category).Error; err != nil {
		fail(w, 409, "category name is already in use")
		return
	}
	respond(w, 201, category)
}
func (s *Server) updateCategory(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	category, ok := s.findCategory(w, chi.URLParam(r, "categoryID"))
	if !ok {
		return
	}
	var input catalogNameInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(120); err != nil {
		fail(w, 400, err.Error())
		return
	}
	category.Name = strings.TrimSpace(input.Name)
	if err := s.db.Save(&category).Error; err != nil {
		fail(w, 409, "category name is already in use")
		return
	}
	respond(w, 200, category)
}
func (s *Server) deleteCategory(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	category, ok := s.findCategory(w, chi.URLParam(r, "categoryID"))
	if !ok {
		return
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("category_id = ?", category.ID).Delete(&models.BookCategory{}).Error; err != nil {
			return err
		}
		return tx.Delete(&category).Error
	}); err != nil {
		internalError(w, err)
		return
	}
	w.WriteHeader(204)
}
func (s *Server) listTags(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	var tags []models.Tag
	if err := s.db.Order("name").Find(&tags).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, tags)
}
func (s *Server) createTag(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var input catalogNameInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(80); err != nil {
		fail(w, 400, err.Error())
		return
	}
	tag := models.Tag{Name: strings.TrimSpace(input.Name)}
	if err := s.db.Create(&tag).Error; err != nil {
		fail(w, 409, "tag name is already in use")
		return
	}
	respond(w, 201, tag)
}
func (s *Server) updateTag(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	tag, ok := s.findTag(w, chi.URLParam(r, "tagID"))
	if !ok {
		return
	}
	var input catalogNameInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(80); err != nil {
		fail(w, 400, err.Error())
		return
	}
	tag.Name = strings.TrimSpace(input.Name)
	if err := s.db.Save(&tag).Error; err != nil {
		fail(w, 409, "tag name is already in use")
		return
	}
	respond(w, 200, tag)
}
func (s *Server) deleteTag(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	tag, ok := s.findTag(w, chi.URLParam(r, "tagID"))
	if !ok {
		return
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tag_id = ?", tag.ID).Delete(&models.BookTag{}).Error; err != nil {
			return err
		}
		return tx.Delete(&tag).Error
	}); err != nil {
		internalError(w, err)
		return
	}
	w.WriteHeader(204)
}
func (s *Server) addCategoryToBook(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	category, ok := s.findCategory(w, chi.URLParam(r, "categoryID"))
	if !ok {
		return
	}
	item := models.BookCategory{BookID: book.ID, CategoryID: category.ID}
	if err := s.db.Where("book_id = ? AND category_id = ?", book.ID, category.ID).FirstOrCreate(&item).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, item)
}
func (s *Server) removeCategoryFromBook(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	category, ok := s.findCategory(w, chi.URLParam(r, "categoryID"))
	if !ok {
		return
	}
	result := s.db.Where("book_id = ? AND category_id = ?", book.ID, category.ID).Delete(&models.BookCategory{})
	if result.Error != nil {
		internalError(w, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		fail(w, 404, "category is not assigned to this book")
		return
	}
	w.WriteHeader(204)
}
func (s *Server) addTagToBook(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	tag, ok := s.findTag(w, chi.URLParam(r, "tagID"))
	if !ok {
		return
	}
	item := models.BookTag{BookID: book.ID, TagID: tag.ID}
	if err := s.db.Where("book_id = ? AND tag_id = ?", book.ID, tag.ID).FirstOrCreate(&item).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, item)
}
func (s *Server) removeTagFromBook(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	tag, ok := s.findTag(w, chi.URLParam(r, "tagID"))
	if !ok {
		return
	}
	result := s.db.Where("book_id = ? AND tag_id = ?", book.ID, tag.ID).Delete(&models.BookTag{})
	if result.Error != nil {
		internalError(w, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		fail(w, 404, "tag is not assigned to this book")
		return
	}
	w.WriteHeader(204)
}
func (s *Server) findCategory(w stdhttp.ResponseWriter, id string) (models.Category, bool) {
	var category models.Category
	if err := s.db.First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(w, 404, "category not found")
		} else {
			internalError(w, err)
		}
		return models.Category{}, false
	}
	return category, true
}
func (s *Server) findTag(w stdhttp.ResponseWriter, id string) (models.Tag, bool) {
	var tag models.Tag
	if err := s.db.First(&tag, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(w, 404, "tag not found")
		} else {
			internalError(w, err)
		}
		return models.Tag{}, false
	}
	return tag, true
}
