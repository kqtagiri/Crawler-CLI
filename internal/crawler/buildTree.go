package crawler

func (crawler *Crawler) BuildTree(pages map[string]*Page, startUrls []string, maxDepth int) []Tree {

	result := []Tree{}

	for _, url := range startUrls {

		currPage, ok := pages[url]
		if !ok {
			continue
		}

		path := map[string]bool{currPage.Url: true}

		var draw func(page *Page, path map[string]bool, depth int) Tree
		draw = func(page *Page, path map[string]bool, depth int) Tree {

			//crawler.Logger.Info("build tree", "url", page.Url, "depth", depth)

			tree := Tree{Resource: page.Url, Title: page.Title, Links: []Tree{}}

			if depth >= maxDepth {
				return tree
			}

			for _, link := range page.Links {

				child, ok := pages[link]
				if !ok || child == nil {
					continue
				}

				if !path[link] {
					path[link] = true
					tree.Links = append(tree.Links, draw(pages[link], path, depth+1))
					delete(path, link)
				}

			}

			return tree

		}

		result = append(result, draw(currPage, path, 0))

	}

	return result

}
