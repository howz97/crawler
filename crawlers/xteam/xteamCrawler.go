package xteam

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/boltdb/bolt"
	"github.com/gocolly/colly"
	"gopkg.in/mgo.v2"
)

const (
	mgoURL              = "localhost:27017"
	bucketName          = "blogUrl"
	boltDB              = "xteam.db"
	lastUrlKey          = "lastUrl"
	secUpdateErr        = "second update can not find Date1 !"
	mongoDBName         = "test"
	mongoCollectionName = "x-teamAnother"
)

var (
	counter  int64 = 1
	tagsArry [10]string
)

type blog struct {
	Title   string   `bson:"title"`
	Author  string   `bson:"author"`
	Date    string   `bson:"date"`
	Photo   string   `bson:"photo"`
	Content string   `bson:"content"`
	Tags    []string `bson:"tags"`
}

type xteamCrawler struct {
	collector       *colly.Collector
	detailCollector *colly.Collector
	boltDB          *bolt.DB
	mongoDB         *mgo.Session
	urlOld          string
	urlNew          string
}

func NewXteam() *xteamCrawler {
	boltDB, err := initBoltDB()
	if err != nil {
		panic(err)
	}

	mgoDB, err := initMgo()
	if err != nil {
		panic(err)
	}

	return &xteamCrawler{
		collector:       colly.NewCollector(),
		detailCollector: colly.NewCollector(),
		boltDB:          boltDB,
		mongoDB:         mgoDB,
	}
}

func initBoltDB() (*bolt.DB, error) {
	return bolt.Open(boltDB, 0600, &bolt.Options{Timeout: 1 * time.Second})
}

func initMgo() (*mgo.Session, error) {
	return mgo.Dial(mgoURL)
}

func (xc *xteamCrawler) closeBoltDB() {
	xc.boltDB.Close()
}

func (xc *xteamCrawler) closeMongoDB() {
	xc.mongoDB.Close()
}

func (xc *xteamCrawler) preUpdate() error {
	errUpdate := xc.boltDB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			log.Fatal(err)
			return err
		}

		lastUrlSli := bucket.Get([]byte(lastUrlKey))
		if lastUrlSli == nil {
			xc.urlOld = "nil"
		} else {
			xc.urlOld = string(lastUrlSli)
		}
		return nil
	})

	return errUpdate
}

func (xc *xteamCrawler) putLastUrlAndExit() {
	errUpdate := xc.boltDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			panic(secUpdateErr)
		}
		err := bucket.Put([]byte(lastUrlKey), []byte(xc.urlNew))
		return err
	})
	if errUpdate != nil {
		log.Fatal(errUpdate)
	} else {
		xc.closeBoltDB()
		xc.closeMongoDB()
		os.Exit(0)
	}
}

func (xc *xteamCrawler) onRequest() {
	xc.collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
}

// traverse a page.
func (xc *xteamCrawler) onHtml() {
	xc.collector.OnHTML("main article h2 a", xc.parse)
}

// traverse a blog.
func (xc *xteamCrawler) detailOnHtml() {
	xc.detailCollector.OnHTML("main article", xc.parseDetail)
}

func (xc *xteamCrawler) visit(url string) {
	err := xc.collector.Visit(url)
	if err != nil {
		if err.Error() == "Not Found" {
			xc.putLastUrlAndExit()
		}
	}
}

func (xc *xteamCrawler) parse(e *colly.HTMLElement) {
	if counter == 1 {
		xc.urlNew = e.Attr("href")
	}
	link := e.Attr("href")
	if link != xc.urlOld {
		xc.detailCollector.Visit("https://x-team.com" + link)
	} else {
		xc.putLastUrlAndExit()
	}

	counter++
}

func (xc *xteamCrawler) parseDetail(e *colly.HTMLElement) {
	title := e.DOM.Find("h1").Text()
	title = strings.TrimSpace(title)

	photo, _ := e.DOM.Find("img").Attr("src")
	author := e.DOM.Find("ul li.post-author-name span[itemprop]").Text()
	date := e.DOM.Find("ul li.post-date span").Text()
	content := e.DOM.Find("section div").Text()
	e.DOM.Find("ul.button-action li ul.option-list li a[title]").Each(func(i int, selection *goquery.Selection) {
		tagsArry[i], _ = selection.Attr("title")
	})
	tags := tagsArry[0:10]

	collection := xc.mongoDB.DB(mongoDBName).C(mongoCollectionName)
	err := collection.Insert(&blog{title, author, date, photo, content, tags})
	if err != nil {
		log.Fatal(err)
	}
}
