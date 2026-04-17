package handler

import (
	"net/http"
)

func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}


	url := r.FormValue("url")

	if url == "" {
		http.Error(
			w,
			"URL is required",
			http.StatusBadRequest,
		)

		return
	}

	w.Write([]byte("Analyze endpoint working"))
}