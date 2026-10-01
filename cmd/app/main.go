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

	resultChan := make(chan *crawler.Page, 1000)
	sch := crawler.NewScheduler(cr, resultChan, maxDepth)

	pages := map[string]*crawler.Page{}
	for range workersCount {

		wg.Add(1)
		go cr.Worker(ctx, sch.Jobs, sch.ResultChan, &sch.Counter, &wg)

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
