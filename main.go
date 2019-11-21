package main

import "fmt"

func main() {
	crawler := NewCrawler(0, nil, nil)
	links, _ := crawler.CrawlWebpage("https://en.wikipedia.org/wiki/Main_Page")
	crawler.CrawlWebpages(links...)

	// Print the keys of the pages the crawler visited
	for key := range crawler.PagesVisited {
		fmt.Println(key)
	}
}
