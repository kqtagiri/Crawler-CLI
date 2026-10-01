package main

import (
	"context"
	"crawler/internal/crawler"
	"encoding/json"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {

	urlFlag := flag.String("urls", "", "list of started urls")
	depthFlag := flag.Int("depth", 1, "max depth")
	timeoutFlag := flag.Duration("timeout", 2*time.Minute, "overall timeout")
	reqTimeoutFlag := flag.Duration("request-timeout", 10*time.Second, "one request timeout")
	outputFlag := flag.String("output", "result.json", "output result file")
	logFlag := flag.String("log", "crawler.log", "log file")

	flag.Parse()

	if *urlFlag == "" {
		log.Fatal("Flag urls cannot be unfilled")
	}

	if *depthFlag < 0 {
		log.Fatal("Flag depth cannot be < 0")
	}

	if *timeoutFlag <= 0 {
		log.Fatal("Flag timeout cannot be <= 0")
	}

	if *reqTimeoutFlag <= 0 {
		log.Fatal("Flag reqTimeout cannot be <= 0")
	}

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
	Logger := slog.New(handler)
	cr := crawler.NewCrawler(*reqTimeoutFlag, Logger)

	startUrls := strings.Split(*urlFlag, ",")
	maxDepth := *depthFlag

	workersCount := 10
	wg := sync.WaitGroup{}
	counter := atomic.Int64{}
	jobs := make(chan crawler.PageDepth, 1000)

	pages := map[string]*crawler.Page{}
	resultChan := make(chan *crawler.Page, 1000)
	for range workersCount {

		wg.Add(1)
		go cr.Worker(ctx, jobs, resultChan, &counter, &wg)

	}

	for _, Url := range startUrls {

		counter.Add(1)
		jobs <- crawler.PageDepth{Url: Url, Depth: 0}

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

	for page := range resultChan {

		pages[page.Url] = page
		//cr.Logger.Info("page stored", "url", page.Url, "depth", page.Depth)
		if page.Depth < maxDepth {
			for _, link := range page.Links {

				if !cr.IsVisited(link) {
					counter.Add(1)
					select {
					case jobs <- crawler.PageDepth{Url: link, Depth: page.Depth + 1}:
					case <-ctx.Done():
						counter.Add(-1)
					}

				}

			}
		}
		counter.Add(-1)

	}

	result := cr.BuildTree(pages, startUrls, maxDepth)

	data, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		cr.Logger.Error("Get next error when create json file", "err", err)
		return
	}

	err = os.WriteFile(*outputFlag, data, 0644)
	if err != nil {
		cr.Logger.Error("Get next error when writing in the file", "err", err)
		return
	}

}
