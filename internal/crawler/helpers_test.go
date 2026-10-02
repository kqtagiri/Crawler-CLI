package crawler

import (
	"io"
	"log/slog"
	"testing"
	"time"
)

func newTestCrawler(t *testing.T) *Crawler {

	t.Helper()
	handler := slog.NewTextHandler(io.Discard, nil)
	logger := slog.New(handler)

	return NewCrawler(time.Second, logger)

}
