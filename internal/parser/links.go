package parser

import (
	"context"
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

var (
	linkCheckClient = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
		},
	}
)

func isLinkAccessible(ctx context.Context, link string) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", link, nil)
	if err != nil {
		return false
	}
	resp, err := linkCheckClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func extractLinks(
	ctx context.Context,
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
				if !isLinkAccessible(ctx, l) {
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
