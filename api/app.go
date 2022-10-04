package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type app struct{}

func (ap *app) handler() http.Handler {
	r := chi.NewRouter()

	r.Get("/", ap.root)
	r.Route("/api", func(r chi.Router) {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Content-Type"},
			AllowCredentials: false,
		}))
		r.Get("/", ap.apiRoot)
	})

	return r
}

func (ap *app) root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func (ap *app) apiRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Hello, API!"}`)
}
