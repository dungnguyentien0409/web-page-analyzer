package parser

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type LinkAnalysisResult struct {
	Internal     int
	External     int
	Inaccessible int
}

func isLinkAccessible(link string) bool {
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Head(link)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func extractLinks(
	doc *goquery.Document,
	baseURLStr string,
) (*LinkAnalysisResult, error) {
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		return nil, err
	}
	res := &LinkAnalysisResult{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	linkSet := make(map[string]bool)
	sem := make(chan struct{}, 20)
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		if href == "" ||
			strings.HasPrefix(href, "#") ||
			strings.HasPrefix(href, "mailto:") ||
			strings.HasPrefix(href, "javascript:") {
			return
		}

		linkURL, err := url.Parse(href)
		if err != nil {
			return
		}

		resolved := baseURL.ResolveReference(linkURL)
		if resolved.Host == "" {
			return
		}
		fullURL := resolved.String()
		if resolved.Host == baseURL.Host {
			res.Internal++
		} else {
			res.External++
		}
		if !linkSet[fullURL] {
			linkSet[fullURL] = true
			wg.Add(1)
			go func(l string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				if !isLinkAccessible(l) {
					mu.Lock()
					res.Inaccessible++
					mu.Unlock()
				}
			}(fullURL)
		}
	})
	wg.Wait()
	return res, nil
}
