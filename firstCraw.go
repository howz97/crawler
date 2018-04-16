package main

import (
	"fmt"

	"github.com/gocolly/colly"
)

func main() {
	// Instantiate default collector
	acollector := colly.NewCollector()

	acollector.OnHTML("div.nav-previous a", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		fmt.Printf(" %s\n", link)
		e.Request.Visit(link)
	})

	// Before making a request print "Visiting ..."
	acollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping on https://hackerspaces.org
	acollector.Visit("https://stackoverflow.blog/company/")
}
