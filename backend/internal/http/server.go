package http

import (
	"errors"
	"log/slog"
	stdhttp "net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"

	"github.com/librearchive/librearchive/internal/config"
	"github.com/librearchive/librearchive/internal/models"
)

type Server struct {
	config      config.Config
	db          *gorm.DB
	rateLimiter *loginRateLimiter
	logger      *slog.Logger
}

func NewServer(cfg config.Config, db *gorm.DB) (*Server, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}
	return &Server{config: cfg, db: db, rateLimiter: newLoginRateLimiter(), logger: slog.Default()}, nil
}

func (s *Server) Router() stdhttp.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(30*time.Second))
	router.Use(s.requestLogger)
	router.Use(s.cors)
	router.Get("/health", s.health)
	router.Get("/ready", s.ready)
	router.Route("/api/v1", func(router chi.Router) {
		router.With(s.limitAuthentication).Post("/auth/bootstrap", s.bootstrap)
		router.With(s.limitAuthentication).Post("/auth/login", s.login)
		router.With(s.limitAuthentication).Post("/auth/refresh", s.refresh)
		router.Group(func(router chi.Router) {
			router.Use(s.authenticate)
			router.Get("/me", s.me)
			router.Post("/auth/logout", s.logout)
			router.Patch("/me", s.updateMyProfile)
			router.Get("/books", s.listBooks)
			router.Get("/categories", s.listCategories)
			router.Get("/tags", s.listTags)
			router.Get("/books/{bookID}", s.getBook)
			router.Get("/books/{bookID}/file", s.serveBook)
			router.Get("/books/{bookID}/cover", s.serveCover)
			router.Get("/books/{bookID}/reading-progress", s.getReadingProgress)
			router.Put("/books/{bookID}/reading-progress", s.upsertReadingProgress)
			router.Get("/reading-progress", s.listReadingProgress)
			router.Get("/favorites", s.listFavorites)
			router.Put("/books/{bookID}/favorite", s.addFavorite)
			router.Delete("/books/{bookID}/favorite", s.removeFavorite)
			router.Get("/collections", s.listCollections)
			router.Post("/collections", s.createCollection)
			router.Get("/collections/{collectionID}", s.getCollection)
			router.Patch("/collections/{collectionID}", s.updateCollection)
			router.Delete("/collections/{collectionID}", s.deleteCollection)
			router.Put("/collections/{collectionID}/books/{bookID}", s.addBookToCollection)
			router.Delete("/collections/{collectionID}/books/{bookID}", s.removeBookFromCollection)
			router.Get("/books/{bookID}/annotations", s.listAnnotations)
			router.Post("/books/{bookID}/highlights", s.createHighlight)
			router.Patch("/highlights/{highlightID}", s.updateHighlight)
			router.Delete("/highlights/{highlightID}", s.deleteHighlight)
			router.Post("/highlights/{highlightID}/notes", s.createNote)
			router.Patch("/notes/{noteID}", s.updateNote)
			router.Delete("/notes/{noteID}", s.deleteNote)
			router.Group(func(router chi.Router) {
				router.Use(requireRole(models.RoleAdmin))
				router.Get("/users", s.listUsers)
				router.Post("/users", s.createUser)
				router.Patch("/users/{userID}", s.updateUser)
				router.Delete("/users/{userID}", s.deleteUser)
				router.Post("/users/{userID}/reset-passphrase", s.resetUserPassphrase)
				router.Post("/books", s.createBook)
				router.Patch("/books/{bookID}", s.updateBook)
				router.Delete("/books/{bookID}", s.deleteBook)
				router.Post("/categories", s.createCategory)
				router.Patch("/categories/{categoryID}", s.updateCategory)
				router.Delete("/categories/{categoryID}", s.deleteCategory)
				router.Post("/tags", s.createTag)
				router.Patch("/tags/{tagID}", s.updateTag)
				router.Delete("/tags/{tagID}", s.deleteTag)
				router.Put("/books/{bookID}/categories/{categoryID}", s.addCategoryToBook)
				router.Delete("/books/{bookID}/categories/{categoryID}", s.removeCategoryFromBook)
				router.Put("/books/{bookID}/tags/{tagID}", s.addTagToBook)
				router.Delete("/books/{bookID}/tags/{tagID}", s.removeTagFromBook)
			})
		})
	})
	return router
}

func (s *Server) health(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	respond(w, stdhttp.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	sqlDB, err := s.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		fail(w, stdhttp.StatusServiceUnavailable, "database is unavailable")
		return
	}
	respond(w, stdhttp.StatusOK, map[string]string{"status": "ready"})
}
