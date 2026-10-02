package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestRunCancelBeforeStart(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	crawler := newTestCrawler(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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

}

func TestRunCancelWhileAddingChildren(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	crawler := newTestCrawler(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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

}
