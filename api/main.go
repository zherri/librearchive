package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/zherri/librearchive/database"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	db := database.Connect()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome"))
	})

	http.ListenAndServe(":3000", r)
}
