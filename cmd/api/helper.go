package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (app *application) readIdParams(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)

	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return int64(id), nil
}
