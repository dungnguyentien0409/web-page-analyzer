package handler

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/dungnguyentien0409/web-page-analyzer/internal/fetcher"
	"github.com/dungnguyentien0409/web-page-analyzer/internal/parser"
)

var tmpl *template.Template
var fetchURL = fetcher.FetchURL

func SetTemplate(t *template.Template) {
	tmpl = t
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w,nil)
}

func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w,"Method not allowed",http.StatusMethodNotAllowed)
		return
	}

	url := r.FormValue("url")

	if url == "" {
		tmpl.Execute(w,map[string]string{
			"Result":"URL is required",
		})
		return
	}

	html, err := fetchURL(url)

	if err != nil {
		tmpl.Execute(w,map[string]string{
			"Result":err.Error(),
		})
		return
	}

	doc, err := parser.ParseHTML(html)

	if err != nil {
		tmpl.Execute(w,map[string]string{
			"Result":"Failed to parse HTML",
		})
		return
	}

	title := parser.ExtractTitle(doc)

	result := fmt.Sprintf("Page title: %s",title)

	tmpl.Execute(w,map[string]string{
		"Result":result,
	})
}