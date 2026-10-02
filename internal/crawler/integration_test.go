package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {

	mtx := sync.Mutex{}
	counts := map[string]int{}

	crawler := newTestCrawler(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		mtx.Lock()
		counts[r.URL.Path]++
		mtx.Unlock()

		switch r.URL.Path {
		case "/":
			w.Write([]byte(`<html><head><title>Home</title></head>
			<body><a href="/B">B</a>
			<a href="/C">C</a>
			</body></html>`))
		case "/B":
			w.Write([]byte(`<html><head><title>B</title></head>
			<body><a href="/">Home</a>
			<a href="/B">B</a>
			<a href="/C">C</a>
			<a href="/D">D</a>
			</body></html>`))
		case "/C":
			w.Write([]byte(`<html><head><title>C</title></head>
			<body><a href="/">Home</a>
			<a href="/B">B</a>
			<a href="/C">C</a>
			<a href="/D">D</a>
			<a href="/E">E</a>
			</body></html>`))
		case "/D":
			w.Write([]byte(`<html><head><title>D</title></head>
			<body><a href="/E">E</a></body></html>`))
		case "/E":
			w.Write([]byte(`<html><head><title>E</title></head>
			<body>no links</body></html>`))
		default:
			w.WriteHeader(404)
		}

	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	startUrls := []string{srv.URL + "/"}
	maxDepth := 3

	workersCount := 10
	workerWg := sync.WaitGroup{}

	resultChan := make(chan *Page, 1000)
	sch := NewScheduler(crawler, resultChan, maxDepth)

	pages := map[string]*Page{}
	for range workersCount {

		workerWg.Add(1)
		go crawler.Worker(ctx, sch.Jobs, sch.ResultChan, &sch.TaskWg, &workerWg)

	}

	schWg := sync.WaitGroup{}
	schWg.Add(1)
	go func() {

		defer schWg.Done()
		sch.Run(ctx, startUrls, pages)

	}()

	workerWg.Wait()
	close(sch.ResultChan)

	schWg.Wait()

	mtx.Lock()
	for path, count := range counts {

		if count != 1 {
			t.Errorf("Get invalid answer: want = 1 request to %s, got = %d", path, count)
		}

	}
	mtx.Unlock()

	want := []Tree{
		{
			Resource: srv.URL + "/",
			Title:    "Home",
			Links: []Tree{
				{
					Resource: srv.URL + "/B",
					Title:    "B",
					Links: []Tree{
						{
							Resource: srv.URL + "/C",
							Title:    "C",
							Links: []Tree{
								{
									Resource: srv.URL + "/D",
									Title:    "D",
									Links:    []Tree{},
								},
								{
									Resource: srv.URL + "/E",
									Title:    "E",
									Links:    []Tree{},
								},
							},
						},
						{
							Resource: srv.URL + "/D",
							Title:    "D",
							Links: []Tree{
								{
									Resource: srv.URL + "/E",
									Title:    "E",
									Links:    []Tree{},
								},
							},
						},
					},
				},
				{
					Resource: srv.URL + "/C",
					Title:    "C",
					Links: []Tree{
						{
							Resource: srv.URL + "/B",
							Title:    "B",
							Links: []Tree{
								{
									Resource: srv.URL + "/D",
									Title:    "D",
									Links:    []Tree{},
								},
							},
						},
						{
							Resource: srv.URL + "/D",
							Title:    "D",
							Links: []Tree{
								{
									Resource: srv.URL + "/E",
									Title:    "E",
									Links:    []Tree{},
								},
							},
						},
						{
							Resource: srv.URL + "/E",
							Title:    "E",
							Links:    []Tree{},
						},
					},
				},
			},
		},
	}

	result := crawler.BuildTree(pages, startUrls, maxDepth)

	if !reflect.DeepEqual(want, result) {
		t.Errorf("want = %+v, got = %+v", want, result)
	}

}
