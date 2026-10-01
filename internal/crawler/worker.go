package crawler

import (
	"context"
	"net/url"
	"sync"
	"sync/atomic"
)

func (crawler *Crawler) Worker(ctx context.Context, jobs chan PageDepth, result chan *Page, counter *atomic.Int64, wg *sync.WaitGroup) {

	defer wg.Done()

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
				counter.Add(-1)
				continue
			}

			base, err := url.Parse(job.Url)
			if err != nil {
				counter.Add(-1)
				continue
			}

			seen := map[string]bool{}
			links := ExtractLinks(nil, seen, doc, base)
			title, _ := ExtractTitle(doc)

			//crawler.Logger.Info("worker result", "url", job.Url, "links", len(links), "title", title)

			result <- &Page{
				Url:    job.Url,
				Title:  title,
				Parent: job.Parent,
				Links:  links,
				Depth:  job.Depth,
			}
		}

	}

}
