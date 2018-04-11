package stackoverflow

import (
	"fmt"
	"log"
)

func Start() {
	crawler := NewStackOverFlow()
	defer crawler.closeBoltDB()
	defer crawler.closeMongoDB()
	err := crawler.preUpdate()
	if err != nil {
		log.Fatal(err)
	}
	crawler.collector.OnHTML("article div.m-post-card__content-column", crawler.parse)
	crawler.detailCollector.OnHTML("article div.m-post-card__content-column", crawler.parseDetail)
	crawler.onRequest()

	for i := 1; i < 170; i++ {
		url := fmt.Sprint("https://stackoverflow.blog/company/page/", i)
		crawler.collector.Visit(url)
	}

	err = crawler.sufUpdate()
	if err != nil {
		log.Fatal(err)
	}
}
