package crawler

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestExtractTitle(t *testing.T) {

	tests := []struct {
		name    string
		htmlStr string
		want    string
		found   bool
	}{
		{
			name:    "simple",
			htmlStr: "<title>X</title>",
			want:    "X",
			found:   true,
		},
		{
			name:    "title in deep",
			htmlStr: "<html><body><title>Hello</title></body></html>",
			want:    "Hello",
			found:   true,
		},
		{
			name:    "no title",
			htmlStr: "<html><body></body></html>",
			want:    "",
			found:   false,
		},
		{
			name:    "empty title",
			htmlStr: "<html><body><title></title></body></html>",
			want:    "",
			found:   true,
		},
		{
			name:    "title with spaces",
			htmlStr: "<html><body><title>  Hello      World   </title></body></html>",
			want:    "Hello World",
			found:   true,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			doc, err := html.Parse(strings.NewReader(tt.htmlStr))
			if err != nil {
				t.Fatalf("html.Parse get next error: %v", err)
			}

			got, found := ExtractTitle(doc)
			if got != tt.want {
				t.Errorf("Get invalid title: want = %s, got = %s", tt.want, got)
			}

			if found != tt.found {
				t.Errorf("Get invalid found: want = %t, got = %t", tt.found, found)
			}

		})

	}

}
