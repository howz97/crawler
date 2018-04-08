package main

import (
	"fmt"
	"strings"
	"github.com/gocolly/colly"

)

func main(){
	c := colly.NewCollector()


	c.OnHTML("li a",func(e *colly.HTMLElement){

        link := e.Attr("href")
        if !(strings.HasSuffix(link,"company/")||strings.HasSuffix(link, "code-for-a-living/")||
        	strings.HasSuffix(link, "insights/")||strings.HasSuffix(link, "engineering/")||
        		strings.HasSuffix(link, "podcasts/")||strings.HasSuffix(link, "nav-developer-hiring-blog")){
        				return
		}

		c.Visit(link)


	})

	c.OnHTML("article div.m-post-card__content-column",func(e *colly.HTMLElement){
		title := e.DOM.Find("h2.m-post-card__title").Find("a").Text()
		link,_ := e.DOM.Find("h2.m-post-card__title").Find("a").Attr("href")
		author := e.DOM.Find("div span.author-name").Text()
		date := e.DOM.Find("div span.date").Text()
		content := e.DOM.Find("div.m-post-card__excerpt").Text()
		fmt.Printf("tltle :%s \n author: %s \n date: %s \n link:%s \n content : %s \n --------------------------------------------------------------------\n ",
			title,author,date,link,content)

	})

	c.OnRequest(func(r *colly.Request){
		fmt.Println("Visiting:", r.URL.String())
	})
	c.Visit("https://stackoverflow.blog")
}
