package crawler

import (
	"testing"
)

func TestIsVisited(t *testing.T) {

	crawler := newTestCrawler(t)

	got := crawler.IsVisited("https://test.com")
	if got != false {
		t.Errorf("Get invalid answer: want = %t, got = %t", false, got)
	}

	got = crawler.IsVisited("https://test.com")
	if got != true {
		t.Errorf("Get invalid answer: want = %t, got = %t", true, got)
	}

}
