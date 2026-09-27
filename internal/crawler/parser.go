package crawler

import (
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
)

var skipExtensions = map[string]bool{
	".tar": true, ".gz": true, ".tgz": true, ".bz2": true, ".xz": true, ".zip": true, ".pkg": true,
	".msi": true, ".exe": true, ".pdf": true, ".jpg": true, ".png": true, ".gif": true,
}

func IsSkippedExt(u *url.URL) bool {
	ext := strings.ToLower(path.Ext(u.Path))
	return skipExtensions[ext]
}

func ExtractTitle(n *html.Node) (string, bool) {

	if n.Type == html.ElementNode && n.Data == "title" {
		title := strings.Join(strings.Fields(GetText(n)), " ")
		return title, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {

		title, found := ExtractTitle(c)
		if found {
			return title, found
		}

	}

	return "", false
}

func GetText(n *html.Node) string {

	if n.Type == html.TextNode {
		return n.Data
	}

	var s strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {

		s.WriteString(GetText(c))

	}

	return s.String()

}

func ExtractLinks(links []string, seen map[string]bool, n *html.Node, base *url.URL) []string {

	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {

			if attr.Key != "href" || attr.Val == "" {
				continue
			}

			if strings.HasPrefix(attr.Val, "#") || strings.HasPrefix(attr.Val, "javascript:") {
				continue
			}

			href, err := url.Parse(attr.Val)
			if err != nil {
				continue
			}

			Url := base.ResolveReference(href)
			if Url.Host != base.Host {
				continue
			}

			if IsSkippedExt(Url) {
				continue
			}

			if seen[Url.String()] {
				continue
			}

			seen[Url.String()] = true
			links = append(links, Url.String())

		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {

		links = ExtractLinks(links, seen, c, base)

	}

	return links

}
