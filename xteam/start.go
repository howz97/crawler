package xteam

import (
	"fmt"
	"log"
)

const (
	site = "https://x-team.com/blog/page/%d/"
)

func (xc *xteamCrawler) Start() {
	err := xc.preUpdate()
	if err != nil {
		log.Fatal(err)
	}
	xc.onRequest()
	xc.onHtml()
	xc.detailOnHtml()
	for pageNumber := 1; true; pageNumber++ {
		url := fmt.Sprintf(site, pageNumber)
		xc.visit(url)
	}
	xc.putLastUrlAndExit()
}
