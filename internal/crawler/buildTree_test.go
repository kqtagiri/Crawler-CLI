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
		want      []Tree
	}{
		{
			name: "basic tree",
			pages: map[string]*Page{
				"https://test.com": {
					Url:    "https://test.com",
					Title:  "test",
					Depth:  0,
					Parent: "",
					Links:  []string{"https://test.com/page1", "https://test.com/page2"},
				},
				"https://test.com/page1": {
					Url:    "https://test.com/page1",
					Title:  "page1",
					Depth:  1,
					Parent: "https://test.com",
					Links:  nil,
				},
				"https://test.com/page2": {
					Url:    "https://test.com/page2",
					Title:  "page2",
					Depth:  1,
					Parent: "https://test.com",
					Links:  nil,
				},
			},
			startUrls: []string{"https://test.com"},
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
			want:      []Tree{},
		},
		{
			name: "loop(a->b->a)",
			pages: map[string]*Page{
				"https://test.com": {
					Url:    "https://test.com",
					Title:  "test",
					Depth:  0,
					Parent: "",
					Links:  []string{"https://test.com/page"},
				},
				"https://test.com/page": {
					Url:    "https://test.com/page",
					Title:  "page",
					Depth:  1,
					Parent: "https://test.com",
					Links:  []string{"https://test.com"},
				},
			},
			startUrls: []string{"https://test.com"},
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
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			crawler := newTestCrawler(t)

			got := crawler.BuildTree(tt.pages, tt.startUrls)
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("Get invalid answer: want = %+v, got = %+v", tt.want, got)
			}

		})

	}

}
