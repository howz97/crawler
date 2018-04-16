package main

import (
	"crawler/crawlers"
	"crawler/crawlers/xteam"
)

func main() {
	var craw crawlers.Crawler
	craw = xteam.NewXteam()
	craw.Start()
}
