package crawler

import (
	"context"
	"sync"
)

type Scheduler struct {
	crawler    *Crawler
	Jobs       chan PageDepth
	ResultChan chan *Page
	MaxDepth   int
	TaskWg     sync.WaitGroup
}

func NewScheduler(crawler *Crawler, resultChan chan *Page, maxDepth int) *Scheduler {

	return &Scheduler{
		crawler:    crawler,
		Jobs:       make(chan PageDepth, 1000),
		ResultChan: resultChan,
		MaxDepth:   maxDepth,
		TaskWg:     sync.WaitGroup{},
	}

}

func (sch *Scheduler) Run(ctx context.Context, startUrls []string, pages map[string]*Page) {

	for _, Url := range startUrls {

		if sch.crawler.IsVisited(Url) {
			continue
		}

		sch.TaskWg.Add(1)
		sch.Jobs <- PageDepth{Url: Url, Depth: 0}

	}

	done := make(chan struct{})
	go func() {

		sch.TaskWg.Wait()
		close(done)

	}()

	for {

		select {
		case page, ok := <-sch.ResultChan:
			if !ok {
				return
			}
			pages[page.Url] = page
			if page.Depth < sch.MaxDepth {
				for _, link := range page.Links {

					if !sch.crawler.IsVisited(link) {
						sch.TaskWg.Add(1)
						select {
						case sch.Jobs <- PageDepth{Url: link, Depth: page.Depth + 1}:
						case <-ctx.Done():
							sch.TaskWg.Done()
							close(sch.Jobs)
							return
						}

					}

				}
			}
			sch.TaskWg.Done()
		case <-done:
			close(sch.Jobs)
			return
		}

	}

}
