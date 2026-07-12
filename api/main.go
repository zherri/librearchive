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

	// "/": AdminLoginPage

	r.Group(func(r chi.Router) {
		r.Use(infra.AuthMiddleware)
		r.Use(infra.AdminOnly)

		// "/register": RegisterPage
		// "/upload": UploadPage
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(infra.AuthMiddleware)
			r.Use(infra.AdminOnly)

			r.Post("/auth/register", authHandler.Register)
			// "/upload": Upload
		})
	})

	http.ListenAndServe(":3000", r)
}
