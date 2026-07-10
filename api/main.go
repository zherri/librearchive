package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/zherri/librearchive/handlers"
	"github.com/zherri/librearchive/infra"
	"github.com/zherri/librearchive/middlewares"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Failed to load .env: %v", err)
	}

	db := infra.ConnectAndMigrate()

	authHandler := handlers.NewAuthHandler(db)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(middlewares.AdminOnly)

			r.Post("/auth/register", authHandler.Register)
		})
	})

	http.ListenAndServe(":3000", r)
}
