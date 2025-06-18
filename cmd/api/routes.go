package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/v1/healthcheck", app.healthcheckHandler)
	// r.Get("/v1/movies", app.moviesHandler)
	// r.Get("/v1/movie{id}", app.movieHandler)

	return r
}
