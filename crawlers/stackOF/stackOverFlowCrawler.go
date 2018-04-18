package stackOF

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gocolly/colly"
	"gopkg.in/mgo.v2"
)

const (
	secUpdateErr        = "second update can not find Date1 !"
	boltDB              = "stackOF.db"
	bucketName          = "blogUrl"
	lastUrlKey          = "lastUrl"
	mgoURL              = "localhost:27017"
	mongoDBName         = "test"
	mongoCollectionName = "stackOverFlowAnother"
)

type blog struct {
	Title   string `bson:"title"`
	Author  string `bson:"author"`
	Date    string `bson:"date"`
	Photo   string `bson:"photo"`
	Content string `bson:"content"`
}

type stackOverFlowCrawler struct {
	collector       *colly.Collector
	detailCollector *colly.Collector
	boltDB          *bolt.DB
	mongoDB         *mgo.Session
	urlNew          string
	urlOld          string
	counter         int64
}

func NewStackOverFlow() *stackOverFlowCrawler {
	boltDB, err := initBoltDB()
	if err != nil {
		panic(err)
	}

	mgoDB, err := initMgo()
	if err != nil {
		panic(err)
	}

	return &stackOverFlowCrawler{
		collector:       colly.NewCollector(),
		detailCollector: colly.NewCollector(),
		boltDB:          boltDB,
		mongoDB:         mgoDB,
		counter:         1,
	}
}

func initBoltDB() (*bolt.DB, error) {
	return bolt.Open(boltDB, 0600, &bolt.Options{Timeout: 1 * time.Second})
}

func (sc *stackOverFlowCrawler) closeBoltDB() {
	err := sc.boltDB.Close()
	if err != nil {
		panic(err)
	}
}

func initMgo() (*mgo.Session, error) {
	return mgo.Dial(mgoURL)
}

func (sc *stackOverFlowCrawler) closeMongoDB() {
	sc.mongoDB.Close()
}

func (sc *stackOverFlowCrawler) preUpdate() error {
	errUpdate := sc.boltDB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		lastUrlSli := bucket.Get([]byte(lastUrlKey))
		if lastUrlSli == nil {
			sc.urlOld = "nil"
			fmt.Println("The first time to crawl.")
		} else {
			sc.urlOld = string(lastUrlSli)
			fmt.Printf("You have crawled by %s last time. \n", sc.urlOld)
		}
		return nil
	})
	return errUpdate
}

func (sc *stackOverFlowCrawler) putLastUrlAndExit() {
	errUpdate := sc.boltDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			panic(secUpdateErr)
		}
		err := bucket.Put([]byte(lastUrlKey), []byte(sc.urlNew))
		return err
	})
	if errUpdate != nil {
		log.Fatal(errUpdate)
	} else {
		sc.closeBoltDB()
		sc.closeMongoDB()
		if sc.urlNew == sc.urlOld {
			fmt.Println("No new blog!")
		}else {
			fmt.Printf("A crawl done, crawl to %s this time. \n", sc.urlNew)
		}
		os.Exit(0)
	}
}

func (sc *stackOverFlowCrawler) parse(e *colly.HTMLElement) {
	if counter == 1 {
		sc.urlNew, _ = e.DOM.Find("h2 a").Attr("href")
	}
	link, _ := e.DOM.Find("h2 a").Attr("href")
	if link != sc.urlOld {
		sc.detailCollector.Visit(link)
	} else {
		sc.putLastUrlAndExit()
	}
	counter++
}

func (sc *stackOverFlowCrawler) parseDetail(e *colly.HTMLElement) {
	title := e.DOM.Find("div.column h1").Text()
	title = strings.TrimSpace(title)
	photo, _ := e.DOM.Find("div span span a img").Attr("src")
	author := e.DOM.Find("div.m-post__meta span.author-name a").Text()
	date := e.DOM.Find("div.m-post__meta span.date time").Text()
	dateNumber, _ := e.DOM.Find("div.m-post__meta span.date time").Attr("datetime")
	dateNumber = strings.TrimRight(dateNumber, "+00:00")
	content := e.DOM.Find("div.m-post-content").Text()
	collection := sc.mongoDB.DB(mongoDBName).C(mongoCollectionName)
	err := collection.Insert(&blog{title, author, date, photo, content})
	if err != nil {
		log.Fatal(err)
	}
}

func (sc *stackOverFlowCrawler) onRequest() {
	sc.collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
}

func (sc *stackOverFlowCrawler) onHtml() {
	sc.collector.OnHTML("article div.m-post-card__content-column", sc.parse)

}

func (sc *stackOverFlowCrawler) detailOnHtml() {
	sc.detailCollector.OnHTML("main#main.site-main", sc.parseDetail)

}

func (sc *stackOverFlowCrawler) visit(url string) {
	err := sc.collector.Visit(url)
	if err != nil {
		if err.Error() == "Not Found" {
			sc.putLastUrlAndExit()
		}
		log.Fatal(err)
	}
}
