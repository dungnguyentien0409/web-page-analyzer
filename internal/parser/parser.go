package parser

import (
	"bytes"
	"io"

	"github.com/PuerkitoBio/goquery"
)

type PageSource struct {
	HTML []byte
	URL  string
}

type PageAnalysis struct {
	HTMLVersion string
	Title       string

	Headings map[string]int

	InternalLinks     int
	ExternalLinks     int
	InaccessibleLinks int

	ContainsLoginForm bool
}

type Analyzer struct {
	documentProvider func(io.Reader) (*goquery.Document, error)
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		documentProvider: goquery.NewDocumentFromReader,
	}
}
func (a *Analyzer) ParseHTML(html []byte) (*goquery.Document, error) {
	reader := bytes.NewReader(html)
	doc, err := a.documentProvider(reader)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
func (a *Analyzer) AnalyzePage(src PageSource) (*PageAnalysis, error) {
	doc, err := a.ParseHTML(src.HTML)
	if err != nil {
		return nil, err
	}
	internal, external, inaccessible, err := extractLinks(doc, src.URL)
	if err != nil {
		return nil, err
	}
	return &PageAnalysis{
		HTMLVersion:       detectHTMLVersion(src.HTML),
		Title:             ExtractTitle(doc),
		Headings:          countHeadings(doc),
		InternalLinks:     internal,
		ExternalLinks:     external,
		InaccessibleLinks: inaccessible,
		ContainsLoginForm: checkLoginForm(doc),
	}, nil
}
