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

type Crawler struct {
	client  *http.Client
	visited map[string]bool
	mtx     sync.Mutex
}

func NewCrawler() *Crawler {

	return &Crawler{
		client:  &http.Client{Timeout: 15 * time.Second},
		visited: map[string]bool{},
		mtx:     sync.Mutex{},
	}

}

func (crawler *Crawler) CreateRequest(ctx context.Context, url string) (*html.Node, error) {

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
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

func (crawler *Crawler) Search(links []string, n *html.Node, base *url.URL) []string {

	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {

			if attr.Key == "href" && attr.Val != "" {
				href, err := url.Parse(attr.Val)
				if err != nil {
					continue
				}

				url := base.ResolveReference(href)
				if url.Host != base.Host {
					continue
				}
				if !crawler.IsVisited(url.String()) {
					links = append(links, url.String())
				}
			}

		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {

		links = crawler.Search(links, c, base)

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

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	crawler := NewCrawler()

	urlFlag := flag.String("urls", "", "list of started urls")
	_ = flag.Int("depth", 1, "max depth")
	_ = flag.Duration("timeout", 2*time.Minute, "overall timeout")
	_ = flag.Duration("request-timeout", 10*time.Second, "one request timeout")
	_ = flag.String("output", "result.json", "output result file")
	_ = flag.String("log", "logs.log", "log file")

	flag.Parse()

	urls := strings.Split(*urlFlag, ",")
	for _, Url := range urls {

		doc, err := crawler.CreateRequest(ctx, Url)
		if err != nil {
			continue
		}

		base, err := url.Parse(Url)
		if err != nil {
			continue
		}

		links := crawler.Search(nil, doc, base)
		for _, link := range links {

			fmt.Println(link)

		}

	}

}
