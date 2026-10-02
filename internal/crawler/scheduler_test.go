package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

func TestRunCancelBeforeStart(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	crawler := newTestCrawler(t)

	counter := atomic.Int64{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		counter.Add(1)

		w.Write([]byte(`<html><head><title>Home</title></head>
		<body><a href="/a">A</a><a href="/b">B</a></body></html>`))

	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	startUrls := []string{srv.URL}
	maxDepth := 1

	workersCount := 10
	wg := sync.WaitGroup{}

	resultChan := make(chan *Page, 1000)
	sch := NewScheduler(crawler, resultChan, maxDepth)

	pages := map[string]*Page{}
	for range workersCount {

		wg.Add(1)
		go crawler.Worker(ctx, sch.Jobs, sch.ResultChan, &sch.TaskWg, &wg)

	}

	cancel()

	schWg := sync.WaitGroup{}
	schWg.Add(1)
	go func() {

		defer schWg.Done()
		sch.Run(ctx, startUrls, pages)

	}()

	wg.Wait()
	close(sch.ResultChan)

	schWg.Wait()

	if got := counter.Load(); got != 0 {
		t.Errorf("Get invalid answer: want = %d requests, got = %d", 0, got)
	}

	if got := len(pages); got != 0 {
		t.Errorf("Get invalid answer: want = %d pages, got = %d", 0, len(pages))
	}

}

func TestRunCancelWhileAddingChildren(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	crawler := newTestCrawler(t)

	rootCounter := atomic.Int64{}
	childCounter := atomic.Int64{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/" {
			rootCounter.Add(1)
		} else {
			childCounter.Add(1)
		}

		w.Write([]byte(`<html><head><title>Home</title></head>
		<body><a href="/a">A</a><a href="/b">B</a></body></html>`))

		cancel()

	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	startUrls := []string{srv.URL}
	maxDepth := 1

	workersCount := 10
	wg := sync.WaitGroup{}

	resultChan := make(chan *Page, 1000)
	sch := NewScheduler(crawler, resultChan, maxDepth)

	pages := map[string]*Page{}
	for range workersCount {

		wg.Add(1)
		go crawler.Worker(ctx, sch.Jobs, sch.ResultChan, &sch.TaskWg, &wg)

	}

	schWg := sync.WaitGroup{}
	schWg.Add(1)
	go func() {

		defer schWg.Done()
		sch.Run(ctx, startUrls, pages)

	}()

	wg.Wait()
	close(sch.ResultChan)

	schWg.Wait()

	if got := rootCounter.Load(); got > 1 {
		t.Errorf("Get invalid answer: want <= %d request, got = %d", 1, got)
	}

	if got := childCounter.Load(); got != 0 {
		t.Errorf("Get invalid answer: want = %d requests, got = %d", 0, got)
	}

}
