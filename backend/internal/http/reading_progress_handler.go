package http

import (
	"errors"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/librearchive/librearchive/internal/models"
)

type readingProgressInput struct {
	Status          models.ReadingStatus `json:"status"`
	CurrentPage     int                  `json:"currentPage"`
	ProgressPercent float64              `json:"progressPercent"`
	IsDownloaded    bool                 `json:"isDownloaded"`
}

func (input readingProgressInput) validate() error {
	if input.Status != models.ReadingStatusUnread && input.Status != models.ReadingStatusInProgress && input.Status != models.ReadingStatusFinished {
		return errors.New("status must be unread, in_progress, or finished")
	}
	if input.CurrentPage < 0 {
		return errors.New("currentPage cannot be negative")
	}
	if input.ProgressPercent < 0 || input.ProgressPercent > 100 {
		return errors.New("progressPercent must be between 0 and 100")
	}
	return nil
}

func (s *Server) getReadingProgress(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	progress := models.ReadingProgress{UserID: currentUser(r).ID, BookID: book.ID, Status: models.ReadingStatusUnread}
	if err := s.db.Preload("Book").Where("user_id = ? AND book_id = ?", progress.UserID, progress.BookID).First(&progress).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		internalError(w, err)
		return
	}
	if progress.Book.ID == 0 {
		progress.Book = book
	}
	respond(w, stdhttp.StatusOK, progress)
}

func (s *Server) upsertReadingProgress(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	book, ok := s.findBook(w, chi.URLParam(r, "bookID"))
	if !ok {
		return
	}
	var input readingProgressInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(); err != nil {
		fail(w, stdhttp.StatusBadRequest, err.Error())
		return
	}
	user := currentUser(r)
	var progress models.ReadingProgress
	err := s.db.Where("user_id = ? AND book_id = ?", user.ID, book.ID).First(&progress).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		progress = models.ReadingProgress{UserID: user.ID, BookID: book.ID}
	} else if err != nil {
		internalError(w, err)
		return
	}
	progress.Status, progress.CurrentPage, progress.ProgressPercent, progress.IsDownloaded = input.Status, input.CurrentPage, input.ProgressPercent, input.IsDownloaded
	if input.Status != models.ReadingStatusUnread {
		now := time.Now().UTC()
		progress.LastReadAt = &now
	}
	if err := s.db.Save(&progress).Error; err != nil {
		internalError(w, err)
		return
	}
	progress.Book = book
	respond(w, stdhttp.StatusOK, progress)
}

func (s *Server) listReadingProgress(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	user := currentUser(r)
	query := s.db.Where("user_id = ?", user.ID)
	if status := strings.TrimSpace(r.URL.Query().Get("status")); status != "" {
		if status != string(models.ReadingStatusUnread) && status != string(models.ReadingStatusInProgress) && status != string(models.ReadingStatusFinished) {
			fail(w, 400, "invalid status filter")
			return
		}
		query = query.Where("status = ?", status)
	}
	if downloaded := r.URL.Query().Get("downloaded"); downloaded != "" {
		if downloaded != "true" && downloaded != "false" {
			fail(w, 400, "downloaded must be true or false")
			return
		}
		query = query.Where("is_downloaded = ?", downloaded == "true")
	}
	offset, limit := pagination(r)
	var total int64
	if err := query.Model(&models.ReadingProgress{}).Count(&total).Error; err != nil {
		internalError(w, err)
		return
	}
	var progress []models.ReadingProgress
	if err := query.Preload("Book").Order("last_read_at desc").Offset(offset).Limit(limit).Find(&progress).Error; err != nil {
		internalError(w, err)
		return
	}
	respond(w, 200, map[string]interface{}{"items": progress, "offset": offset, "limit": limit, "total": total})
}
