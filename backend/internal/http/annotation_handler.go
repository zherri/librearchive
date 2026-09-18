package http

import (
	"errors"
	stdhttp "net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/librearchive/librearchive/internal/models"
	"gorm.io/gorm"
)

type highlightInput struct {
	SelectedText string `json:"selectedText"`
	Color        string `json:"color"`
	Page         int    `json:"page"`
	StartOffset  *int   `json:"startOffset"`
	EndOffset    *int   `json:"endOffset"`
	ChapterRef   string `json:"chapterRef"`
}

func (input highlightInput) validate() error {
	if strings.TrimSpace(input.SelectedText) == "" {
		return errors.New("selectedText is required")
	}
	if input.Page < 0 {
		return errors.New("page cannot be negative")
	}
	if strings.TrimSpace(input.Color) == "" {
		return errors.New("color is required")
	}
	if input.StartOffset != nil && *input.StartOffset < 0 {
		return errors.New("startOffset cannot be negative")
	}
	if input.EndOffset != nil && *input.EndOffset < 0 {
		return errors.New("endOffset cannot be negative")
	}
	if input.StartOffset != nil && input.EndOffset != nil && *input.EndOffset < *input.StartOffset {
		return errors.New("endOffset must not precede startOffset")
	}
	return nil
}
func (s *Server) listAnnotations(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	user := currentUser(r)
	kind := r.URL.Query().Get("type")
	color := strings.TrimSpace(r.URL.Query().Get("color"))
	if kind != "" && kind != "highlight" && kind != "note" {
		fail(w, 400, "type must be highlight or note")
		return
	}
	if kind == "note" {
		query := s.db.Where("notes.user_id = ? AND notes.book_id = ?", user.ID, book.ID)
		if color != "" {
			query = query.Joins("JOIN highlights ON highlights.id = notes.highlight_id").Where("highlights.color = ?", color)
		}
		var notes []models.Note
		if err := query.Order("notes.created_at desc").Find(&notes).Error; err != nil {
			internalError(w, err)
			return
		}
		respond(w, 200, map[string]interface{}{"type": "note", "items": notes})
		return
	}
	query := s.db.Where("user_id = ? AND book_id = ?", user.ID, book.ID)
	if color != "" {
		query = query.Where("color = ?", color)
	}
	var highlights []models.Highlight
	if err := query.Preload("Notes", "user_id = ?", user.ID).Order("created_at desc").Find(&highlights).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, map[string]interface{}{"type": "highlight", "items": highlights})
}
func (s *Server) createHighlight(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	var input highlightInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(); err != nil {
		fail(w, 400, err.Error())
		return
	}
	highlight := models.Highlight{UserID: currentUser(r).ID, BookID: book.ID, SelectedText: strings.TrimSpace(input.SelectedText), Color: strings.TrimSpace(input.Color), Page: input.Page, StartOffset: input.StartOffset, EndOffset: input.EndOffset, ChapterRef: strings.TrimSpace(input.ChapterRef)}
	if err := s.db.Create(&highlight).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 201, highlight)
}
func (s *Server) updateHighlight(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	highlight, ok := s.findOwnedHighlight(w, r)
	if !ok {
		return
	}
	var input highlightInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(); err != nil {
		fail(w, 400, err.Error())
		return
	}
	highlight.SelectedText, highlight.Color, highlight.Page, highlight.StartOffset, highlight.EndOffset, highlight.ChapterRef = strings.TrimSpace(input.SelectedText), strings.TrimSpace(input.Color), input.Page, input.StartOffset, input.EndOffset, strings.TrimSpace(input.ChapterRef)
	if err := s.db.Save(&highlight).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, highlight)
}
func (s *Server) deleteHighlight(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	highlight, ok := s.findOwnedHighlight(w, r)
	if !ok {
		return
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("highlight_id = ?", highlight.ID).Delete(&models.Note{}).Error; err != nil {
			return err
		}
		return tx.Delete(&highlight).Error
	}); err != nil {
		internalError(w, err)
		return
	}
	w.WriteHeader(204)
}

type noteInput struct {
	Content string `json:"content"`
}

func (input noteInput) validate() error {
	if strings.TrimSpace(input.Content) == "" {
		return errors.New("content is required")
	}
	return nil
}
func (s *Server) createNote(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	highlight, ok := s.findOwnedHighlight(w, r)
	if !ok {
		return
	}
	var input noteInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(); err != nil {
		fail(w, 400, err.Error())
		return
	}
	note := models.Note{UserID: currentUser(r).ID, BookID: highlight.BookID, HighlightID: highlight.ID, Content: strings.TrimSpace(input.Content)}
	if err := s.db.Create(&note).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 201, note)
}
func (s *Server) updateNote(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	note, ok := s.findOwnedNote(w, r)
	if !ok {
		return
	}
	var input noteInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(); err != nil {
		fail(w, 400, err.Error())
		return
	}
	note.Content = strings.TrimSpace(input.Content)
	if err := s.db.Save(&note).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, note)
}
func (s *Server) deleteNote(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	note, ok := s.findOwnedNote(w, r)
	if !ok {
		return
	}
	if err := s.db.Delete(&note).Error; err != nil {
		internalError(w, err)
		return
	}
	w.WriteHeader(204)
}
func (s *Server) findOwnedHighlight(w stdhttp.ResponseWriter, r *stdhttp.Request) (models.Highlight, bool) {
	var highlight models.Highlight
	if err := s.db.Where("id = ? AND user_id = ?", chi.URLParam(r, "highlightID"), currentUser(r).ID).First(&highlight).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(w, 404, "highlight not found")
		} else {
			internalError(w, err)
		}
		return models.Highlight{}, false
	}
	return highlight, true
}
func (s *Server) findOwnedNote(w stdhttp.ResponseWriter, r *stdhttp.Request) (models.Note, bool) {
	var note models.Note
	if err := s.db.Where("id = ? AND user_id = ?", chi.URLParam(r, "noteID"), currentUser(r).ID).First(&note).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(w, 404, "note not found")
		} else {
			internalError(w, err)
		}
		return models.Note{}, false
	}
	return note, true
}
