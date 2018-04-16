package crawlers

// every crawler has Start().
type Crawler interface {
	Start()
}

// start a crawler from here
func StartCrawler(cb Crawler) {
	cb.Start()
}
