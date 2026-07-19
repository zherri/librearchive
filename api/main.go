package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/zherri/librearchive/handlers"
	"github.com/zherri/librearchive/infra"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Failed to load .env: %v", err)
	}

	db := infra.ConnectAndMigrate()

	infra.CreateAdminUser(db)

	authHandler := handlers.NewAuthHandler(db)
	bookHandler := handlers.NewBookHandler(db)
	panelHandler := handlers.NewPanelHandler(db)

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(infra.AuthMiddleware)
			r.Use(infra.AdminOnly)

			r.Get("/panel/overview", panelHandler.Overview)
			r.Post("/auth/register", authHandler.Register)
			r.Post("/book/upload", bookHandler.Upload)
		})

		r.Group(func(r chi.Router) {
			r.Use(infra.AuthMiddleware)

			r.Get("/book/get", bookHandler.Get)
			r.Get("/book/get/{id}", bookHandler.GetByID)
			r.Patch("/book/update/{id}", bookHandler.Update)
			r.Delete("/book/delete/{id}", bookHandler.Delete)
		})
	})

	http.ListenAndServe(":3000", r)
}
