package parser

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

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
) (int, int, int, error) {
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		return 0, 0, 0, err
	}
	internal := 0
	external := 0
	inaccessible := 0
	var wg sync.WaitGroup
	var mu sync.Mutex
	linkSet := make(map[string]bool)

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
			internal++
		} else {
			external++
		}

		if !linkSet[fullURL] {
			linkSet[fullURL] = true
			wg.Add(1)
			go func(l string) {
				defer wg.Done()
				if !isLinkAccessible(l) {
					mu.Lock()
					inaccessible++
					mu.Unlock()
				}
			}(fullURL)
		}
	})
	wg.Wait()
	return internal, external, inaccessible, nil
}
