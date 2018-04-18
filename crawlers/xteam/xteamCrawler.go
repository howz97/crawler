package xteam

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/boltdb/bolt"
	"github.com/gocolly/colly"
	"gopkg.in/mgo.v2"
)

const (
	secUpdateErr        = "second update can not find Date1 !"
	boltDB              = "xteam.db"
	bucketName          = "blogUrl"
	urlOldKey           = "lastUrl"
	mgoURL              = "localhost:27017"
	mongoDBName         = "test"
	mongoCollectionName = "x-team"
)

var (
	counter int64 = 1
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
	counter         int64
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
		counter:         1,
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
			return err
		}

		urlOldSli := bucket.Get([]byte(urlOldKey))
		if urlOldSli == nil {
			xc.urlOld = "nil"
			fmt.Println("The first time to crawl.")
		} else {
			xc.urlOld = string(urlOldSli)
			fmt.Printf("You have crawled by %s last time. \n", xc.urlOld)
		}
		return nil
	})
	return errUpdate
}

// xc.sufUpdate is the last method called , exit directly in it.
func (xc *xteamCrawler) putLastUrlAndExit() {
	errUpdate := xc.boltDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return errors.New(secUpdateErr)
		}
		err := bucket.Put([]byte(urlOldKey), []byte(xc.urlNew))
		return err
	})
	if errUpdate != nil {
		log.Fatal(errUpdate)
	} else {
		xc.closeBoltDB()
		xc.closeMongoDB()
		if xc.urlNew == xc.urlOld {
			fmt.Println("No new blog!")
		}else {
			fmt.Printf("A crawl done, crawl to %s this time. \n", xc.urlNew)
		}
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
		log.Fatal(err)
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
	var blog = &blog{}

	blog.Title = e.DOM.Find("h1").Text()
	blog.Title = strings.TrimSpace(blog.Title)
	blog.Photo, _ = e.DOM.Find("img").Attr("src")
	blog.Author = e.DOM.Find("ul li.post-author-name span[itemprop]").Text()
	blog.Date = e.DOM.Find("ul li.post-date span").Text()
	blog.Content = e.DOM.Find("section div").Text()
	e.DOM.Find("ul.button-action li ul.option-list li a[title]").Each(func(i int, selection *goquery.Selection) {
		tag, _ := selection.Attr("title")
		blog.Tags = append(blog.Tags, tag)
	})

	collection := xc.mongoDB.DB(mongoDBName).C(mongoCollectionName)
	err := collection.Insert(blog)
	if err != nil {
		log.Fatal(err)
	}
}
