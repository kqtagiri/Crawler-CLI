package crawler

import (
	"context"
	"sync/atomic"
	"time"
)

type Scheduler struct {
	crawler    *Crawler
	Jobs       chan PageDepth
	ResultChan chan *Page
	MaxDepth   int
	Counter    atomic.Int64
}

func NewScheduler(crawler *Crawler, resultChan chan *Page, maxDepth int) *Scheduler {

	return &Scheduler{
		crawler:    crawler,
		Jobs:       make(chan PageDepth, 1000),
		ResultChan: resultChan,
		MaxDepth:   maxDepth,
		Counter:    atomic.Int64{},
	}

}

func (sch *Scheduler) Run(ctx context.Context, startUrls []string, pages map[string]*Page) {

	for _, Url := range startUrls {

		sch.Counter.Add(1)
		sch.Jobs <- PageDepth{Url: Url, Depth: 0}

	}

	done := make(chan struct{})
	go func() {

		defer close(done)

		for {

			if sch.Counter.Load() <= 0 {
				close(sch.Jobs)
				return
			}

			select {
			case <-ctx.Done():
				close(sch.Jobs)
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}

		}

	}()

	for page := range sch.ResultChan {

		pages[page.Url] = page
		if page.Depth < sch.MaxDepth {
			for _, link := range page.Links {

				if !sch.crawler.IsVisited(link) {
					sch.Counter.Add(1)
					select {
					case sch.Jobs <- PageDepth{Url: link, Depth: page.Depth + 1}:
					case <-ctx.Done():
						continue
					}

				}

			}
		}

	}

	<-done

}
