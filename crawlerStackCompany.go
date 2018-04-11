package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

func main() {

	c := colly.NewCollector()
	detailCollector := c.Clone()

	c.OnHTML("article div.m-post-card__content-column", func(e *colly.HTMLElement) {

		link, _ := e.DOM.Find("h2 a").Attr("href")
		detailCollector.Visit(link)

	})

	detailCollector.OnHTML("main#main.site-main", func(e *colly.HTMLElement) {

		title := e.DOM.Find("div.column h1").Text()
		title = strings.TrimSpace(title)
		photo, _ := e.DOM.Find("div span span a img").Attr("src")
		author := e.DOM.Find("div.m-post__meta span.author-name a").Text()
		date := e.DOM.Find("div.m-post__meta span.date time").Text()
		content := e.DOM.Find("div.m-post-content").Text()
		fmt.Printf("<<%s>> \n%s     on %s \nphoto :%s \n%s\n----------------------------------------------------\n",
			title, author, date, photo, content)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	for i := 1; i < 170; i++ {
		url := fmt.Sprint("https://stackoverflow.blog/company/page/", i)

		c.Visit(url)
	}
}
