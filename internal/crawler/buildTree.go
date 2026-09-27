package crawler

func (crawler *Crawler) BuildTree(pages map[string]*Page, startUrls []string) []Tree {

	result := []Tree{}

	for _, url := range startUrls {

		currPage, ok := pages[url]
		if !ok {
			continue
		}

		var draw func(page *Page, depth int) Tree
		draw = func(page *Page, depth int) Tree {

			crawler.Logger.Info("build tree", "url", page.Url, "depth", depth)

			tree := Tree{Resource: page.Url, Title: page.Title, Links: []Tree{}}
			for _, link := range page.Links {

				child, ok := pages[link]
				if !ok {
					continue
				}

				if child.Parent == page.Url {
					tree.Links = append(tree.Links, draw(pages[link], depth+1))
				}

			}

			return tree

		}

		result = append(result, draw(currPage, 0))

	}

	return result

}
