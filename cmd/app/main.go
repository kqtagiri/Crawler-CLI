package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/html"
)

type Page struct {
	Url   string
	Title string
	Links []string
}

func CreateRequest(ctx context.Context, client *http.Client, url string) (*html.Node, error) {

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("Error when creating new request:", err)
		return nil, err
	}

	resp, err := client.Do(req)
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

func search(links []string, n *html.Node, base *url.URL) []string {

	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {

			if attr.Key == "href" && attr.Val != "" {
				href, err := url.Parse(attr.Val)
				if err != nil {
					continue
				}

				url := base.ResolveReference(href).String()
				links = append(links, url)
			}

		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {

		links = search(links, c, base)

	}

	return links

}

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	urlFlag := flag.String("urls", "", "list of started urls")
	_ = flag.Int("depth", 1, "max depth")
	_ = flag.Duration("timeout", 2*time.Minute, "overall timeout")
	_ = flag.Duration("request-timeout", 10*time.Second, "one request timeout")
	_ = flag.String("output", "result.json", "output result file")
	_ = flag.String("log", "logs.log", "log file")

	flag.Parse()

	urls := strings.Split(*urlFlag, ",")
	for _, Url := range urls {

		doc, err := CreateRequest(ctx, client, Url)
		if err != nil {
			continue
		}

		base, err := url.Parse(Url)
		if err != nil {
			continue
		}

		links := search(nil, doc, base)
		for _, link := range links {

			fmt.Println(link)

		}

	}

}
