package main

import "fmt"

func main() {
	originalUrL := "https://en.wikipedia.org/wiki/Main_Page"
	var iterations uint32 = 250

	crawler := NewCrawler(1, nil, nil)
	visited, unvisited := crawler.CrawlNRecursively(originalUrL, iterations)

	fmt.Printf("***********************************\n\n")
	fmt.Printf("The Original Url was:\n")
	fmt.Printf("%s\n\n", originalUrL)
	fmt.Printf("After %d iterations:\n", iterations)
	fmt.Printf("%8d: URLs were Visited, and\n", len(visited))
	fmt.Printf("%8d: URLs are Unvisited.\n\n", len(unvisited))
	fmt.Printf("***********************************\n")
}
