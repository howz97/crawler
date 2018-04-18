package stackOF

import (
	"fmt"
	"log"
)

const (
	site = "https://stackoverflow.blog/company/page/"
)

func (sc *stackOverFlowCrawler) Start() {
	err := sc.preUpdate()
	if err != nil {
		log.Fatal(err)
	}
	sc.onRequest()
	sc.onHtml()
	sc.detailOnHtml()
	pageNumber := 1
	for {
		url := fmt.Sprint(site, pageNumber)
		sc.visit(url)
		pageNumber++
	}
	sc.putLastUrlAndExit()
}
