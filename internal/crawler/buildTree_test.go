package crawler

import (
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"
)

func newTestCrawler(t *testing.T) *Crawler {

	t.Helper()
	handler := slog.NewTextHandler(io.Discard, nil)
	logger := slog.New(handler)

	return NewCrawler(time.Second, logger)

}

func TestBuildTree(t *testing.T) {

	tests := []struct {
		name      string
		pages     map[string]*Page
		startUrls []string
		maxDepth  int
		want      []Tree
	}{
		{
			name: "basic tree",
			pages: map[string]*Page{
				"https://test.com": {
					Url:   "https://test.com",
					Title: "test",
					Depth: 0,
					Links: []string{"https://test.com/page1", "https://test.com/page2"},
				},
				"https://test.com/page1": {
					Url:   "https://test.com/page1",
					Title: "page1",
					Depth: 1,
					Links: nil,
				},
				"https://test.com/page2": {
					Url:   "https://test.com/page2",
					Title: "page2",
					Depth: 1,
					Links: nil,
				},
			},
			startUrls: []string{"https://test.com"},
			maxDepth:  2,
			want: []Tree{
				{
					Resource: "https://test.com",
					Title:    "test",
					Links: []Tree{
						{
							Resource: "https://test.com/page1",
							Title:    "page1",
							Links:    []Tree{},
						},
						{
							Resource: "https://test.com/page2",
							Title:    "page2",
							Links:    []Tree{},
						},
					},
				},
			},
		},
		{
			name:      "empty start urls",
			pages:     map[string]*Page{},
			startUrls: []string{},
			maxDepth:  1,
			want:      []Tree{},
		},
		{
			name: "loop(A->B->A)",
			pages: map[string]*Page{
				"https://test.com": {
					Url:   "https://test.com",
					Title: "test",
					Depth: 0,
					Links: []string{"https://test.com/page"},
				},
				"https://test.com/page": {
					Url:   "https://test.com/page",
					Title: "page",
					Depth: 1,
					Links: []string{"https://test.com"},
				},
			},
			startUrls: []string{"https://test.com"},
			maxDepth:  2,
			want: []Tree{
				{
					Resource: "https://test.com",
					Title:    "test",
					Links: []Tree{
						{
							Resource: "https://test.com/page",
							Title:    "page",
							Links:    []Tree{},
						},
					},
				},
			},
		},
		{
			name: "loop(A->A)",
			pages: map[string]*Page{
				"https://test.com": {
					Url:   "https://test.com",
					Title: "test",
					Depth: 0,
					Links: []string{"https://test.com"},
				},
			},
			startUrls: []string{"https://test.com"},
			maxDepth:  2,
			want: []Tree{
				{
					Resource: "https://test.com",
					Title:    "test",
					Links:    []Tree{},
				},
			},
		},
		{
			name: "graph",
			pages: map[string]*Page{
				"A": {
					Url:   "A",
					Title: "A",
					Depth: 0,
					Links: []string{"B", "C"},
				},
				"B": {
					Url:   "B",
					Title: "B",
					Depth: 1,
					Links: []string{"A", "B", "C", "D"},
				},
				"C": {
					Url:   "C",
					Title: "C",
					Depth: 1,
					Links: []string{"A", "B", "C", "D", "E"},
				},
				"D": {
					Url:   "D",
					Title: "D",
					Depth: 2,
					Links: []string{"E"},
				},
				"E": {
					Url:   "E",
					Title: "E",
					Depth: 2,
					Links: []string{},
				},
			},
			startUrls: []string{"A"},
			maxDepth:  3,
			want: []Tree{
				{
					Resource: "A",
					Title:    "A",
					Links: []Tree{
						{
							Resource: "B",
							Title:    "B",
							Links: []Tree{
								{
									Resource: "C",
									Title:    "C",
									Links: []Tree{
										{
											Resource: "D",
											Title:    "D",
											Links:    []Tree{},
										},
										{
											Resource: "E",
											Title:    "E",
											Links:    []Tree{},
										},
									},
								},
								{
									Resource: "D",
									Title:    "D",
									Links: []Tree{
										{
											Resource: "E",
											Title:    "E",
											Links:    []Tree{},
										},
									},
								},
							},
						},
						{
							Resource: "C",
							Title:    "C",
							Links: []Tree{
								{
									Resource: "B",
									Title:    "B",
									Links: []Tree{
										{
											Resource: "D",
											Title:    "D",
											Links:    []Tree{},
										},
									},
								},
								{
									Resource: "D",
									Title:    "D",
									Links: []Tree{
										{
											Resource: "E",
											Title:    "E",
											Links:    []Tree{},
										},
									},
								},
								{
									Resource: "E",
									Title:    "E",
									Links:    []Tree{},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			crawler := newTestCrawler(t)

			got := crawler.BuildTree(tt.pages, tt.startUrls, tt.maxDepth)
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("Get invalid answer: want = %+v, got = %+v", tt.want, got)
			}

		})

	}

}
