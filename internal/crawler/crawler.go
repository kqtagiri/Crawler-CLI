package crawler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

func (crawler *Crawler) IsVisited(url string) bool {

	crawler.mtx.Lock()
	defer crawler.mtx.Unlock()
	if !crawler.visited[url] {
		crawler.visited[url] = true
		return false
	}
	return true

}

func (crawler *Crawler) CreateRequest(ctx context.Context, url string) (*html.Node, error) {

	reqCtx, cancel := context.WithTimeout(ctx, crawler.reqTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		crawler.Logger.Error("Get next error when creating new request", "err", err)
		return nil, err
	}

	req.Header.Set("User-Agent", "Crawler-CLI/1.0 (+https://github.com/kqtagiri/Crawler-CLI)")

	resp, err := crawler.client.Do(req)
	if err != nil {
		crawler.Logger.Error("Get next error when doing request", "err", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		crawler.Logger.Error("Get bad status", "status", resp.Status)
		return nil, fmt.Errorf("Bad status: %s", resp.Status)
	}

	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/xhtml") && !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		crawler.Logger.Warn("Get not html content-type", "url", url)
		return nil, fmt.Errorf("Not html content-type: %s", url)
	}

	crawler.Logger.Info("", "url", url, "Content-Type", resp.Header.Get("Content-Type"))

	doc, err := html.Parse(resp.Body)
	if err != nil {
		crawler.Logger.Error("Get next error when parsing html", "err", err)
		return nil, err
	}

	return doc, nil

}
