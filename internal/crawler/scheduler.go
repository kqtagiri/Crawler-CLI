package crawler

import (
	"context"
)

type Scheduler struct {
	Jobs    chan PageDepth
	AddJobs chan PageDepth
}

func NewScheduler() *Scheduler {

	return &Scheduler{
		Jobs:    make(chan PageDepth, 1000),
		AddJobs: make(chan PageDepth, 1000),
	}

}

func (sch *Scheduler) Run(ctx context.Context, startUrls []string) {

	for _, Url := range startUrls {

		sch.Jobs <- PageDepth{Url: Url, Depth: 0}

	}

	defer close(sch.Jobs)

	for job := range sch.AddJobs {

		select {
		case sch.Jobs <- PageDepth{Url: job.Url, Depth: job.Depth}:
		case <-ctx.Done():

		}

	}

}
