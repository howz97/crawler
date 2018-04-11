package stackoverflow

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gocolly/colly"
	"gopkg.in/mgo.v2"
	"log"
	"strconv"
	"strings"
	"time"
)

var lastDate int

type blog struct {
	Title   string `bson:"title"`
	Author  string `bson:"author"`
	Date    string `bson:"date"`
	Photo   string `bson:"photo"`
	Content string `bson:"content"`
}

type StackOverFlowCrawler struct {
	collector       *colly.Collector
	detailCollector *colly.Collector
	boltDB          *bolt.DB
	mongoDB         *mgo.Session
}

func NewStackOverFlow() *StackOverFlowCrawler {

	boltdb, errOpen := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if errOpen != nil {
		log.Fatal(errOpen)
	}
	session, errDial := mgo.Dial("localhost:27017")
	if errDial != nil {
		log.Fatal(errDial)
	}

	return &StackOverFlowCrawler{
		collector:       colly.NewCollector(),
		detailCollector: colly.NewCollector(),
		boltDB:          boltdb,
		mongoDB:         session,
	}
}

func (sc *StackOverFlowCrawler) closeBoltDB() {
	sc.boltDB.Close()
}

func (sc *StackOverFlowCrawler) closeMongoDB() {
	sc.mongoDB.Close()
}

func (sc *StackOverFlowCrawler) preUpdate() error {
	errUpdate := sc.boltDB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("Date1"))
		if err != nil {
			log.Fatal(err)
			return err
		}

		lastDateSli := bucket.Get([]byte("lastDate"))
		if lastDateSli == nil {
			lastDate = 0
		} else {
			lastDate, err = strconv.Atoi(string(lastDateSli))
			if err != nil {
				log.Fatal(err)
				return err
			}
		}

		return nil
	})
	if errUpdate != nil {
		return errUpdate
	} else {
		return nil
	}
}

func (sc *StackOverFlowCrawler) sufUpdate() error {
	errUpdate := sc.boltDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Date1"))
		if bucket == nil {
			panic("second update can not find Date1 !")
		}
		s := strconv.Itoa(lastDate)
		err := bucket.Put([]byte("lastDate"), []byte(s))
		return err
	})
	if errUpdate != nil {
		return errUpdate
	} else {
		return nil
	}
}

func (sc *StackOverFlowCrawler) parse(e *colly.HTMLElement) {

	link, _ := e.DOM.Find("h2 a").Attr("href")
	sc.detailCollector.Visit(link)

}

func (sc *StackOverFlowCrawler) parseDetail(e *colly.HTMLElement) {

	title := e.DOM.Find("div.column h1").Text()
	title = strings.TrimSpace(title)
	photo, _ := e.DOM.Find("div span span a img").Attr("src")
	author := e.DOM.Find("div.m-post__meta span.author-name a").Text()
	date := e.DOM.Find("div.m-post__meta span.date time").Text()
	dateNumber, _ := e.DOM.Find("div.m-post__meta span.date time").Attr("datetime")
	dateNumber = strings.TrimRight(dateNumber, "+00:00")
	content := e.DOM.Find("div.m-post-content").Text()

	c := sc.mongoDB.DB("test").C("crawlerStackOverFlow2")
	if checkDate(dateNumber, &lastDate) {
		err := c.Insert(&blog{title, author, date, photo, content})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (sc *StackOverFlowCrawler) onRequest() {
	sc.collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
}
