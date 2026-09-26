package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/html"
)

type Page struct {
	Url   string
	Title string
	Links []string
}

type PageDepth struct {
	Url   string
	Depth int
}

type Crawler struct {
	client     *http.Client
	visited    map[string]bool
	mtx        sync.Mutex
	reqTimeout time.Duration
}

func NewCrawler(reqTimeout time.Duration) *Crawler {

	return &Crawler{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		},
		visited:    map[string]bool{},
		mtx:        sync.Mutex{},
		reqTimeout: reqTimeout,
	}

}

func (crawler *Crawler) CreateRequest(ctx context.Context, url string) (*html.Node, error) {

	reqCtx, cancel := context.WithTimeout(ctx, crawler.reqTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("Error when creating new request:", err)
		return nil, err
	}

	resp, err := crawler.client.Do(req)
	if err != nil {
		fmt.Println("Error doing request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Get bad status:", resp.Status)
		return nil, fmt.Errorf("Bad status: %s", resp.Status)
	}

	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/xhtml") && !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		fmt.Println("Get bad url, not html")
		return nil, fmt.Errorf("Bad url, not html: %s", url)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Println("Error parsing html:", err)
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
		return false
	}
	return true

}

func main() {

	urlFlag := flag.String("urls", "", "list of started urls")
	depthFlag := flag.Int("depth", 1, "max depth")
	timeoutFlag := flag.Duration("timeout", 2*time.Minute, "overall timeout")
	reqTimeoutFlag := flag.Duration("request-timeout", 10*time.Second, "one request timeout")
	_ = flag.String("output", "result.json", "output result file")
	_ = flag.String("log", "logs.log", "log file")

	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithTimeout(ctx, *timeoutFlag)
	defer cancel()

	crawler := NewCrawler(*reqTimeoutFlag)

	urls := strings.Split(*urlFlag, ",")
	maxDepth := *depthFlag

	queue := []PageDepth{}
	for _, Url := range urls {

		queue = append(queue, PageDepth{Url: Url, Depth: 0})

	}

	for len(queue) > 0 {

		Job := queue[0]
		queue = queue[1:]

		if crawler.IsVisited(Job.Url) {
			continue
		}
		crawler.visited[Job.Url] = true

		doc, err := crawler.CreateRequest(ctx, Job.Url)
		if err != nil {
			continue
		}

		base, err := url.Parse(Job.Url)
		if err != nil {
			continue
		}

		seen := map[string]bool{}
		links := crawler.ExtractLinks(nil, seen, doc, base)
		title, _ := ExtractTitle(doc)
		fmt.Printf("\n\nURL:%s\nTitle:%s\nLinks:\n", Job.Url, title)
		for _, link := range links {

			fmt.Println("\t", link)
			if Job.Depth < maxDepth {
				if !crawler.IsVisited(link) {
					queue = append(queue, PageDepth{Url: link, Depth: Job.Depth + 1})
				}
			}

		}
	}

}
