package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// make data
	data := map[string]string{
		"status":    "available",
		"available": app.config.env,
		"version":   version,
	}

	// parse and check error
	jsonData, err := json.Marshal(data)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Failed parsing data to json", http.StatusInternalServerError)
		return
	}

	// set header and send response
	w.Header().Set("Content-Type", "applicaton/json")
	w.Write(jsonData)
}
