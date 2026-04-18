package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/handler"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/parser"
)

func main() {
	tmpl := template.Must(
		template.ParseFiles("web/templates/index.html"),
	)

	analyzer := parser.NewAnalyzer()
	h := handler.NewHandler(tmpl, analyzer)

	http.HandleFunc("/", h.IndexHandler)
	http.HandleFunc("/analyze", h.AnalyzeHandler)

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("web/static")),
		),
	)

	log.Println("Server running at http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal(err)
	}
}
