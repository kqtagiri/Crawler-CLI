package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestCreateRequest(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><body>test</body></html>`))

		})

		srv := httptest.NewServer(handler)
		defer srv.Close()

		crawler := newTestCrawler(t)
		doc, err := crawler.CreateRequest(context.Background(), srv.URL)

		if err != nil {
			t.Fatalf("Get unexpected error: %v", err)
		}

		if doc == nil {
			t.Fatalf("Get invalid doc")
		}

	})

	t.Run("Not html", func(t *testing.T) {

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`<html><body>test</body></html>`))

		})

		srv := httptest.NewServer(handler)
		defer srv.Close()

		crawler := newTestCrawler(t)
		_, err := crawler.CreateRequest(context.Background(), srv.URL)

		if err == nil {
			t.Fatalf("Wait error, get nil")
		}

	})

	t.Run("Timeout", func(t *testing.T) {

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			time.Sleep(2 * time.Second)

			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><body>test</body></html>`))

		})

		srv := httptest.NewServer(handler)
		defer srv.Close()

		crawler := newTestCrawler(t)
		_, err := crawler.CreateRequest(context.Background(), srv.URL)

		if err == nil {
			t.Fatalf("Wait error, get nil")
		}

	})

	t.Run("Skip redirect", func(t *testing.T) {

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			http.Redirect(w, r, "https://test.com", 301)

		})

		srv := httptest.NewServer(handler)
		defer srv.Close()

		crawler := newTestCrawler(t)
		_, err := crawler.CreateRequest(context.Background(), srv.URL)

		if err == nil {
			t.Fatalf("Wait error, get nil")
		}

	})

	t.Run("Not found", func(t *testing.T) {

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(404)
			w.Write([]byte(`<html><body>test</body></html>`))

		})

		srv := httptest.NewServer(handler)
		defer srv.Close()

		crawler := newTestCrawler(t)
		_, err := crawler.CreateRequest(context.Background(), srv.URL)

		if err == nil {
			t.Fatalf("Wait error, get nil")
		}

	})

}
