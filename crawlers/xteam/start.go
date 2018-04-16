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

	// traverse blog
	pageNumber := 1
	for {
		url := fmt.Sprintf(site, pageNumber)
		xc.visit(url)
		pageNumber++
	}

	// xc.sufUpdate is the last method called , exit directly in it.
	// So do not add anything below.
	xc.putLastUrlAndExit()
}
