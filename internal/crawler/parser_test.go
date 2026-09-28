package crawler

import (
	"net/url"
	"reflect"
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

func TestExtractLinks(t *testing.T) {

	tests := []struct {
		name    string
		links   []string
		seen    map[string]bool
		htmlStr string
		baseUrl string
		want    []string
	}{
		{
			name:    "working",
			links:   nil,
			seen:    map[string]bool{},
			htmlStr: `<a href="/page1"></a><a href="/page2"></a>`,
			baseUrl: "https://test.com",
			want:    []string{"https://test.com/page1", "https://test.com/page2"},
		},
		{
			name:    "incorrect html node type + incorrect data",
			links:   nil,
			seen:    map[string]bool{},
			htmlStr: `<!-- comment --><p>test</p>`,
			baseUrl: "https://test.com",
			want:    nil,
		},
		{
			name:    "incorrect attr key",
			links:   nil,
			seen:    map[string]bool{},
			htmlStr: `<a class="nav">test</a>`,
			baseUrl: "https://test.com",
			want:    nil,
		},
		{
			name:    "empty attr val",
			links:   nil,
			seen:    map[string]bool{},
			htmlStr: `<a href="">test</a>`,
			baseUrl: "https://test.com",
			want:    nil,
		},
		{
			name:    "prefix #",
			links:   nil,
			seen:    map[string]bool{},
			htmlStr: `<a href="#test">test</a>`,
			baseUrl: "https://test.com",
			want:    nil,
		},
		{
			name:    "prefix javascript",
			links:   nil,
			seen:    map[string]bool{},
			htmlStr: `<a href="javascript:">test</a>`,
			baseUrl: "https://test.com",
			want:    nil,
		},
		{
			name:    "different hosts",
			links:   nil,
			seen:    map[string]bool{},
			htmlStr: `<a href="https://other.com/page">test</a>`,
			baseUrl: "https://test.com",
			want:    nil,
		},
		{
			name:    "skip extension",
			links:   nil,
			seen:    map[string]bool{},
			htmlStr: `<a href="/file.pdf">pdf</a><a href="/page.html">html</a>`,
			baseUrl: "https://test.com",
			want:    []string{"https://test.com/page.html"},
		},
		{
			name:    "duplicate in seen",
			links:   nil,
			seen:    map[string]bool{"https://test.com/page1": true},
			htmlStr: `<a href="/page1"></a>`,
			baseUrl: "https://test.com",
			want:    nil,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			doc, err := html.Parse(strings.NewReader(tt.htmlStr))
			if err != nil {
				t.Fatalf("html.Parse get next error: %v", err)
			}

			base, err := url.Parse(tt.baseUrl)
			if err != nil {
				t.Fatalf("url.Parse get next error: %v", err)
			}

			got := ExtractLinks(tt.links, tt.seen, doc, base)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get invalid answer: want = %v, got = %v", tt.want, got)
			}

		})

	}

}

func TestGetText(t *testing.T) {

	tests := []struct {
		name    string
		htmlStr string
		want    string
	}{
		{
			name:    "simple TextNode",
			htmlStr: `title`,
			want:    "title",
		},
		{
			name:    "empty text",
			htmlStr: `<p></p>`,
			want:    "",
		},
		{
			name:    "nested text",
			htmlStr: `<p><b>Hello </b><b>World</b></p>`,
			want:    "Hello World",
		},
		{
			name:    "text in siblings",
			htmlStr: `<div><p>Hello </p><p>World</p></div>`,
			want:    "Hello World",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			doc, err := html.Parse(strings.NewReader(tt.htmlStr))
			if err != nil {
				t.Fatalf("html.Parse get next error: %v", err)
			}

			got := GetText(doc)
			if tt.want != got {
				t.Errorf("Get invalid answer: want = %v, got = %v", tt.want, got)
			}

		})

	}

}
