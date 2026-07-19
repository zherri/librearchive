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

	ah := handlers.NewAuthHandler(db)
	bh := handlers.NewBookHandler(db)
	ph := handlers.NewPanelHandler(db)

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/login", ah.Login)

		r.Group(func(r chi.Router) {
			r.Use(infra.AuthMiddleware)
			r.Use(infra.AdminOnly)

			r.Get("/panel/overview", ph.Overview)
			r.Post("/auth/register", ah.Register)
			r.Post("/book/upload", bh.Upload)
			r.Patch("/book/update/{id}", bh.Update)
			r.Delete("/book/delete/{id}", bh.Delete)
		})

		r.Group(func(r chi.Router) {
			r.Use(infra.AuthMiddleware)

			r.Get("/book/get", bh.Get)
			r.Get("/book/get/{id}", bh.GetByID)
		})
	})

	http.ListenAndServe(":3000", r)
}
