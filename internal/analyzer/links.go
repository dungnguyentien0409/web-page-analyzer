package analyzer

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

func (a *DefaultAnalyzer) isLinkAccessible(ctx context.Context, link string) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", link, nil)
	if err != nil {
		return false
	}
	a.logger.Debug("checking link", "url", link)
	var resp *http.Response
	for i := 0; i < a.retryCount; i++ {
		resp, err = linkCheckClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		if ctx.Err() != nil {
			return false
		}
		a.logger.Warn("retrying link check", "url", link, "attempt", i+1)
	}
	if err != nil {
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func (a *DefaultAnalyzer) extractLinks(
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
	jobs := make(chan string)
	for w := 0; w < a.workerCount; w++ {
		go func() {
			for l := range jobs {
				if !a.isLinkAccessible(ctx, l) {
					mu.Lock()
					res.Inaccessible++
					mu.Unlock()
				}
				wg.Done()
			}
		}()
	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
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

		resolved.Fragment = ""
		fullURL := resolved.String()
		if !linkSet[fullURL] {
			linkSet[fullURL] = true
			if resolved.Host == baseURL.Host {
				res.Internal++
			} else {
				res.External++
			}
			wg.Add(1)
			jobs <- fullURL
		}
	})
	close(jobs)
	wg.Wait()
	return res, nil
}
