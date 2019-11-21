package main

func main() {
	crawler := NewCrawler(0, nil, nil)
	links := []string{
		"https://en.wikipedia.org/wiki/Main_Page",
		"https://www.reddit.com/r/all/",
	}
	crawler.CrawlWebpages(links...)
	//fmt.Printf("%+v\n", crawler.PagesVisited)
}
