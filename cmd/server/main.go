package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/handler"
)

func main() {
	tmpl := template.Must(
		template.ParseFiles("web/templates/index.html"),
	)

	handler.SetTemplate(tmpl)

	http.HandleFunc("/", handler.IndexHandler)
	http.HandleFunc("/analyze", handler.AnalyzeHandler)

	log.Println("Server running at http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal(err)
	}
}