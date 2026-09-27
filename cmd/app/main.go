package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/net/html"
)

type Page struct {
	Url    string
	Title  string
	Depth  int
	Parent string
	Links  []string
}

type PageDepth struct {
	Url    string
	Depth  int
	Parent string
}

type Tree struct {
	Resource string `json:"resource"`
	Title    string `json:"title"`
	Links    []Tree `json:"links"`
}

type Crawler struct {
	client     *http.Client
	visited    map[string]bool
	mtx        sync.Mutex
	reqTimeout time.Duration
	logger     *slog.Logger
}

var skipExtensions = map[string]bool{
	".tar": true, ".gz": true, ".tgz": true, ".bz2": true, ".xz": true, ".zip": true, ".pkg": true,
	".msi": true, ".exe": true, ".pdf": true, ".jpg": true, ".png": true, ".gif": true,
}

func IsSkippedExt(u *url.URL) bool {
	ext := strings.ToLower(path.Ext(u.Path))
	return skipExtensions[ext]
}

func NewCrawler(reqTimeout time.Duration, logger *slog.Logger) *Crawler {

	return &Crawler{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		},
		visited:    map[string]bool{},
		mtx:        sync.Mutex{},
		reqTimeout: reqTimeout,
		logger:     logger,
	}

}

func (crawler *Crawler) CreateRequest(ctx context.Context, url string) (*html.Node, error) {

	reqCtx, cancel := context.WithTimeout(ctx, crawler.reqTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		crawler.logger.Error("Get next error when creating new request", "err", err)
		return nil, err
	}

	resp, err := crawler.client.Do(req)
	if err != nil {
		crawler.logger.Error("Get next error when doing request", "err", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		crawler.logger.Error("Get bad status", "status", resp.Status)
		return nil, fmt.Errorf("Bad status: %s", resp.Status)
	}

	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/xhtml") && !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		crawler.logger.Warn("Get bad url, not html", "url", url)
		return nil, fmt.Errorf("Bad url, not html: %s", url)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		crawler.logger.Error("Get next error when parsing html", "err", err)
		return nil, err
	}

	return doc, nil

}

func ExtractTitle(n *html.Node) (string, bool) {

	if n.Type == html.ElementNode && n.Data == "title" {
		title := strings.Join(strings.Fields(GetText(n)), " ")
		return title, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {

		title, found := ExtractTitle(c)
		if found {
			return title, found
		}

	}

	return "", false
}

func GetText(n *html.Node) string {

	if n.Type == html.TextNode {
		return n.Data
	}

	var s strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {

		s.WriteString(GetText(c))

	}

	return s.String()

}

func (crawler *Crawler) ExtractLinks(links []string, seen map[string]bool, n *html.Node, base *url.URL) []string {

	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {

			if attr.Key != "href" || attr.Val == "" {
				continue
			}

			if strings.HasPrefix(attr.Val, "#") || strings.HasPrefix(attr.Val, "javascript:") {
				continue
			}

			href, err := url.Parse(attr.Val)
			if err != nil {
				continue
			}

			Url := base.ResolveReference(href)
			if Url.Host != base.Host {
				continue
			}

			if IsSkippedExt(Url) {
				continue
			}

			if seen[Url.String()] {
				continue
			}

			seen[Url.String()] = true
			links = append(links, Url.String())

		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {

		links = crawler.ExtractLinks(links, seen, c, base)

	}

	return links

}

func (crawler *Crawler) IsVisited(url string) bool {

	crawler.mtx.Lock()
	defer crawler.mtx.Unlock()
	if !crawler.visited[url] {
		crawler.visited[url] = true
		return false
	}
	return true

}

func (crawler *Crawler) BuildTree(pages map[string]*Page, startUrls []string) []Tree {

	result := []Tree{}

	for _, url := range startUrls {

		currPage, ok := pages[url]
		if !ok {
			continue
		}

		var draw func(page *Page, depth int) Tree
		draw = func(page *Page, depth int) Tree {

			crawler.logger.Info("build tree", "url", page.Url, "depth", depth)

			tree := Tree{Resource: page.Url, Title: page.Title, Links: []Tree{}}
			for _, link := range page.Links {

				child, ok := pages[link]
				if !ok {
					continue
				}

				if child.Parent == page.Url {
					tree.Links = append(tree.Links, draw(pages[link], depth+1))
				}

			}

			return tree

		}

		result = append(result, draw(currPage, 0))

	}

	return result

}

func (crawler *Crawler) Worker(ctx context.Context, jobs chan PageDepth, result chan *Page, counter *atomic.Int64, wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			doc, err := crawler.CreateRequest(ctx, job.Url)
			if err != nil {
				counter.Add(-1)
				continue
			}

			base, err := url.Parse(job.Url)
			if err != nil {
				counter.Add(-1)
				continue
			}

			seen := map[string]bool{}
			links := crawler.ExtractLinks(nil, seen, doc, base)
			title, _ := ExtractTitle(doc)

			crawler.logger.Info("worker result", "url", job.Url, "links", len(links), "title", title)

			result <- &Page{
				Url:    job.Url,
				Title:  title,
				Parent: job.Parent,
				Links:  links,
				Depth:  job.Depth,
			}
		}

	}

}

func main() {

	urlFlag := flag.String("urls", "", "list of started urls")
	depthFlag := flag.Int("depth", 1, "max depth")
	timeoutFlag := flag.Duration("timeout", 2*time.Minute, "overall timeout")
	reqTimeoutFlag := flag.Duration("request-timeout", 10*time.Second, "one request timeout")
	outputFlag := flag.String("output", "result.json", "output result file")
	logFlag := flag.String("log", "crawler.log", "log file")

	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithTimeout(ctx, *timeoutFlag)
	defer cancel()

	file, err := os.OpenFile(*logFlag, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Get next error when open log file:", err)
	}
	defer file.Close()

	handler := slog.NewTextHandler(file, nil)
	logger := slog.New(handler)
	crawler := NewCrawler(*reqTimeoutFlag, logger)

	startUrls := strings.Split(*urlFlag, ",")
	maxDepth := *depthFlag

	workersCount := 10
	wg := sync.WaitGroup{}
	counter := atomic.Int64{}
	jobs := make(chan PageDepth, 1000)

	pages := map[string]*Page{}
	resultChan := make(chan *Page, 1000)
	for range workersCount {

		wg.Add(1)
		go crawler.Worker(ctx, jobs, resultChan, &counter, &wg)

	}

	go func() {

		for counter.Load() > 0 {
			if ctx.Err() != nil {
				close(jobs)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}

		close(jobs)

	}()

	go func() {

		wg.Wait()
		close(resultChan)

	}()

	for _, Url := range startUrls {

		counter.Add(1)
		jobs <- PageDepth{Url: Url, Depth: 0, Parent: ""}

	}

	for page := range resultChan {

		pages[page.Url] = page
		crawler.logger.Info("page stored", "url", page.Url, "depth", page.Depth, "parent", page.Parent)
		if page.Depth < maxDepth {
			for _, link := range page.Links {

				if !crawler.IsVisited(link) {
					counter.Add(1)
					select {
					case jobs <- PageDepth{Url: link, Depth: page.Depth + 1, Parent: page.Url}:
					case <-ctx.Done():
						counter.Add(-1)
					}

				}

			}
		}
		counter.Add(-1)

	}

	result := crawler.BuildTree(pages, startUrls)

	data, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		crawler.logger.Error("Get next error when create json file", "err", err)
		return
	}

	err = os.WriteFile(*outputFlag, data, 0644)
	if err != nil {
		crawler.logger.Error("Get next error when writing in the file", "err", err)
		return
	}

}
