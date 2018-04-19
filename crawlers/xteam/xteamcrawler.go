// Copyright 2018 The zh1014. All rights reserved.
package xteam

import (
	"errors"
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
	boltDBName          = "xteam.db"
	bucketName          = "blogUrl"
	mgoURL              = "localhost:27017"
	mongoDBName         = "crawlerDB"
	mongoCollectionName = "x-team"
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
	lastUrl         string
	newestUrl       string
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
		counter:         0,
	}
}

func initBoltDB() (*bolt.DB, error) {
	return bolt.Open(boltDBName, 0600, &bolt.Options{Timeout: 1 * time.Second})
}

func initMgo() (*mgo.Session, error) {
	return mgo.Dial(mgoURL)
}

func (xc *xteamCrawler) closeBoltDB() {
	err := xc.boltDB.Close()
	if err != nil {
		panic(err)
	}
}

func (xc *xteamCrawler) closeMongoDB() {
	xc.mongoDB.Close()
}

func (xc *xteamCrawler) preUpdate() error {
	return xc.boltDB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		lastUrlSlice := bucket.Get([]byte("lastUrl"))
		if lastUrlSlice == nil {
			xc.lastUrl = ""
			fmt.Println("The first time to crawl.")
		} else {
			xc.lastUrl = string(lastUrlSlice)
			fmt.Printf("You have crawled by %s last time. \n", xc.lastUrl)
		}
		return nil
	})
}

// putLastUrlAndExit is the last method called , exit directly in it.
func (xc *xteamCrawler) putLastUrlAndExit() {
	errUpdate := xc.boltDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return errors.New("can not find bucket")
		}
		err := bucket.Put([]byte("lastUrl"), []byte(xc.newestUrl))
		return err
	})
	if errUpdate != nil {
		xc.closeBoltDB()
		xc.closeMongoDB()
		log.Fatal(errUpdate)
	} else {
		xc.closeBoltDB()
		xc.closeMongoDB()
		if xc.newestUrl == xc.lastUrl {
			fmt.Println("No new blog!")
		} else {
			fmt.Printf("once crawl finished, crawl to %s this time. \n", xc.newestUrl)
		}
		os.Exit(0)
	}
}

func (xc *xteamCrawler) visit(url string) {
	err := xc.collector.Visit(url)
	if err != nil {
		if err.Error() == "Not Found" {
			xc.putLastUrlAndExit()
		}
		xc.closeBoltDB()
		xc.closeMongoDB()
		log.Fatal(err)
	}
}

func (xc *xteamCrawler) onRequest() {
	xc.collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
}

func (xc *xteamCrawler) onHtml() {
	xc.collector.OnHTML("main article h2 a", xc.parse)
}

func (xc *xteamCrawler) detailOnHtml() {
	xc.detailCollector.OnHTML("main article", xc.parseDetail)
}

func (xc *xteamCrawler) parse(e *colly.HTMLElement) {
	if xc.counter == 0 {
		xc.newestUrl = e.Attr("href")
	}
	link := e.Attr("href")
	if link != xc.lastUrl {
		xc.detailCollector.Visit("https://x-team.com" + link)
	} else {
		xc.putLastUrlAndExit()
	}
	xc.counter++
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
		xc.closeBoltDB()
		xc.closeMongoDB()
		log.Fatal(err)
	}
}
