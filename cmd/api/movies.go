package main

import (
	"fmt"
	"net/http"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParams(r)
	if err != nil || id < 1 {
		http.Error(w, "Invalid id", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Show the details of the movie %v", id)
}