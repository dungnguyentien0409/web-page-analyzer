package main

import (
	"log"
	"net/http"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/handler"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(
			w,
			r,
			"web/templates/index.html",
		)

	})

	http.HandleFunc(
		"/analyze",
		handler.AnalyzeHandler,
	)

	log.Println("Server running at http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}

}