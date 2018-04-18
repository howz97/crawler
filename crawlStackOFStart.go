package main

import (
	"crawler/crawlers"
	"crawler/crawlers/stackOF"
)

func main() {
	var craw crawlers.Crawler
	craw = stackOF.NewStackOverFlow()
	crawlers.StartCrawler(craw)
}
