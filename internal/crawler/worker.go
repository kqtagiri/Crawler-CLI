package crawler

import (
	"context"
	"net/url"
	"sync"
)

func (crawler *Crawler) Worker(ctx context.Context, jobs <-chan PageDepth, result chan<- *Page, taskWg *sync.WaitGroup, workerWg *sync.WaitGroup) {

	defer workerWg.Done()

	for {

		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			doc, err := crawler.CreateRequest(ctx, job.Url)
			if err != nil {
				taskWg.Done()
				continue
			}

			base, err := url.Parse(job.Url)
			if err != nil {
				taskWg.Done()
				continue
			}

			seen := map[string]bool{}
			links := ExtractLinks(nil, seen, doc, base)
			title, _ := ExtractTitle(doc)

			select {
			case result <- &Page{
				Url:   job.Url,
				Title: title,
				Links: links,
				Depth: job.Depth,
			}:
			case <-ctx.Done():
				taskWg.Done()
				return
			}

		}

	}

}
