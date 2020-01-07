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
	site = "https://stackoverflow.article/company/page/"

	boltDBName      = "stackOverFlow"
	boltBucket      = "myBucket"
	progressURL     = "progressURL"
	mongoDSN        = "localhost:27017"
	mongoDBName     = "stackOverFlow"
	mongoCollection = "myCollection"
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
	mongo            *mgo.Session
	start            string
	destination      string
	wg				 *sync.WaitGroup
}

func New() *Crawler {
	blt, err := bolt.Open(boltDBName, os.ModePerm, bolt.DefaultOptions)
	if err != nil {
		panic(err)
	}

	mongo, err := mgo.Dial(mongoDSN)
	if err != nil {
		panic(err)
	}

	crawler := &Crawler{
		pageCollector:    colly.NewCollector(),
		articleCollector: colly.NewCollector(),
		bolt:             blt,
		mongo:            mongo,
		wg:				  &sync.WaitGroup{},
	}
	err = crawler.getProgress()
	if err != nil {
		panic(err)
	}
	crawler.articleCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("requesting article: ", r.URL.String())
	})
	crawler.pageCollector.OnHTML("article div.m-post-card__content-column", crawler.handleArticleLink)
	crawler.articleCollector.OnHTML("main#main.site-main", crawler.saveArticle)
	return crawler
}

func (sc *Crawler) Run() {
	for pageNumber := 1; true ; pageNumber++ {
		url := fmt.Sprint(site, pageNumber)
		err := sc.pageCollector.Visit(url)
		if err != nil {
			if strings.Contains(err.Error(), "Not Found") {
				sc.UpdateProgressAndExit()
			}
			log.Println(err)
		}
	}
}

func (sc *Crawler)getProgress() error {
	return sc.bolt.View(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(boltBucket))
		if err != nil {
			return err
		}
		progress := bucket.Get([]byte(progressURL))
		sc.destination = ""
		if progress != nil {
			sc.destination = string(progress)
		}
		return nil
	})
}

func (sc *Crawler) UpdateProgressAndExit() {
	sc.wg.Wait()
	sc.mongo.Close()
	err := sc.bolt.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(boltBucket))
		if bucket == nil {
			panic(fmt.Sprintf("bucket %v not exists", boltBucket))
		}
		return bucket.Put([]byte(progressURL), []byte(sc.start))
	})
	if err != nil {
		log.Println(err)
	}
	err = sc.bolt.Close()
	if err != nil {
		log.Println(err)
	}
	os.Exit(0)
}

func (sc *Crawler) handleArticleLink(e *colly.HTMLElement) {
	link, exists := e.DOM.Find("h2 a").Attr("href")
	if !exists {
		log.Println("no link in this HTML element")
	}
	if sc.start == "" {
		sc.start = link
	}
	if link != sc.destination {
		sc.wg.Add(1)
		go sc.articleCollector.Visit(link)
	} else {
		sc.UpdateProgressAndExit()
	}
}

func (sc *Crawler) saveArticle(e *colly.HTMLElement) {
	title := e.DOM.Find("div.column h1").Text()
	title = strings.TrimSpace(title)
	photo, _ := e.DOM.Find("div span span a img").Attr("src")
	author := e.DOM.Find("div.m-post__meta span.author-name a").Text()
	date := e.DOM.Find("div.m-post__meta span.date time").Text()
	dateNumber, _ := e.DOM.Find("div.m-post__meta span.date time").Attr("datetime")
	dateNumber = strings.TrimRight(dateNumber, "+00:00")
	content := e.DOM.Find("div.m-post-content").Text()
	collection := sc.mongo.DB(mongoDBName).C(mongoCollection)
	err := collection.Insert(&article{title, author, date, photo, content})
	if err != nil {
		log.Println("saveArticle: ", err)
	}
	sc.wg.Done()
}
