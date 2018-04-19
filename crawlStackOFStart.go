package main

import (
	"crawler/crawlers"
	"crawler/crawlers/stackoverflow"
)

func main() {
	var craw crawlers.Crawler
	craw = stackoverflow.NewStackOverFlow()
	craw.Start()
}
