// Copyright 2018 The zh1014. All rights reserved.
package stackoverflow

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gocolly/colly"
	"gopkg.in/mgo.v2"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	site = "https://stackoverflow.blog/company/page/"

	boltDBName      = "stackOverFlow"
	boltBucket      = "myBucket"
	progressURL     = "progressURL"
	mongoDBName     = "stackOverFlow"
	mongoCollection = "myCollection"
	mongoDSN        = "mongodb://localhost:27017/" + mongoDBName
)

type article struct {
	Title   string `bson:"title"`
	Author  string `bson:"author"`
	Date    string `bson:"date"`
	Photo   string `bson:"photo"`
	Content string `bson:"content"`
}

type Crawler struct {
	pageCollector    *colly.Collector
	articleCollector *colly.Collector
	bolt             *bolt.DB
	mgoS             *mgo.Session
	start            string
	destination      string
	wg               *sync.WaitGroup
}

func New() *Crawler {
	blt, err := bolt.Open(boltDBName, os.ModePerm, bolt.DefaultOptions)
	if err != nil {
		panic(err)
	}

	mongoSession, err := mgo.Dial(mongoDSN)
	if err != nil {
		panic(err)
	}

	crawler := &Crawler{
		pageCollector:    colly.NewCollector(),
		articleCollector: colly.NewCollector(),
		bolt:             blt,
		mgoS:             mongoSession,
		wg:               &sync.WaitGroup{},
	}
	err = crawler.getProgress()
	if err != nil {
		panic(err)
	}
	crawler.pageCollector.OnRequest(func(r *colly.Request) {
		log.Println("requesting page: ", r.URL.String())
	})
	crawler.articleCollector.OnRequest(func(r *colly.Request) {
		log.Println("requesting article: ", r.URL.String())
	})
	crawler.pageCollector.OnHTML("article div.m-post-card__content-column", crawler.handleArticleLink)
	crawler.articleCollector.OnHTML("main#main.site-main", crawler.saveArticle)
	return crawler
}

func (cler *Crawler) Run() {
	for pageNumber := 1; true; pageNumber++ {
		url := fmt.Sprint(site, pageNumber)
		err := cler.pageCollector.Visit(url)
		if err != nil {
			if strings.Contains(err.Error(), "Not Found") {
				cler.UpdateProgressAndExit()
			}
			log.Println(err)
		}
	}
}

func (cler *Crawler) getProgress() error {
	return cler.bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(boltBucket))
		if err != nil {
			return err
		}
		progress := bucket.Get([]byte(progressURL))
		cler.destination = ""
		if progress != nil {
			cler.destination = string(progress)
		}
		return nil
	})
}

func (cler *Crawler) UpdateProgressAndExit() {
	cler.wg.Wait()
	cler.mgoS.Close()
	err := cler.bolt.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(boltBucket))
		if bucket == nil {
			panic(fmt.Sprintf("bucket %v not exists", boltBucket))
		}
		return bucket.Put([]byte(progressURL), []byte(cler.start))
	})
	if err != nil {
		fmt.Println(err)
	}
	err = cler.bolt.Close()
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(0)
}

func (cler *Crawler) handleArticleLink(e *colly.HTMLElement) {
	link, exists := e.DOM.Find("h2 a").Attr("href")
	if !exists {
		fmt.Println("[ERROR]no link in this HTML element")
	}
	if cler.start == "" {
		cler.start = link
	}
	if link != cler.destination {
		cler.wg.Add(1)
		go cler.visitArticle(link) // request, parse and store one article in a separate goroutine
	} else {
		cler.UpdateProgressAndExit()
	}
}

// visitArticle request, parse and store an article in a separate goroutine
func (cler *Crawler) visitArticle(url string) {
	err := cler.articleCollector.Visit(url)
	if err != nil {
		fmt.Println("[ERROR]visit article "+url+":", err.Error())
	}
	cler.wg.Done()
}

func (cler *Crawler) saveArticle(e *colly.HTMLElement) {
	title := e.DOM.Find("div.column h1").Text()
	title = strings.TrimSpace(title)
	photo, _ := e.DOM.Find("div span span a img").Attr("src")
	author := e.DOM.Find("div.m-post__meta span.author-name a").Text()
	date := e.DOM.Find("div.m-post__meta span.date time").Text()
	content := e.DOM.Find("div.m-post-content").Text()
	collection := cler.mgoS.DB(mongoDBName).C(mongoCollection)
	err := collection.Insert(&article{title, author, date, photo, content})
	if err != nil {
		fmt.Println("[ERROR]saveArticle: ", err)
	}
}
