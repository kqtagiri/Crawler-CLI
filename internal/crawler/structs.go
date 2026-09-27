package crawler

import (
	"log/slog"
	"net/http"
	"sync"
	"time"
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
	Logger     *slog.Logger
}

func NewCrawler(reqTimeout time.Duration, Logger *slog.Logger) *Crawler {

	return &Crawler{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		},
		visited:    map[string]bool{},
		mtx:        sync.Mutex{},
		reqTimeout: reqTimeout,
		Logger:     Logger,
	}

}
