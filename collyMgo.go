package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gocolly/colly"
	"gopkg.in/mgo.v2"
)

type blog struct {
	Title   string `bson:"title"`
	Author  string `bson:"author"`
	Date    string `bson:"date"`
	Photo   string `bson:"photo"`
	Content string `bson:"content"`
}

var (
	acount     int64 = 1
	atemporary int
)

func main() {
	var lastDate int

	db, errOpen := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if errOpen != nil {
		log.Fatal(errOpen)
	}
	defer db.Close()

	//connect to mongodb server
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	errUpdate := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("Date1"))
		if err != nil {
			log.Fatal("create bucket error ")
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

	c := colly.NewCollector()
	detailCollector := c.Clone()

	c.OnHTML("article div.m-post-card__content-column", func(e *colly.HTMLElement) {

		link, _ := e.DOM.Find("h2 a").Attr("href")
		detailCollector.Visit(link)

	})

	detailCollector.OnHTML("main#main.site-main", func(e *colly.HTMLElement) {

		title := e.DOM.Find("div.column h1").Text()
		title = strings.TrimSpace(title)
		photo, _ := e.DOM.Find("div span span a img").Attr("src")
		author := e.DOM.Find("div.m-post__meta span.author-name a").Text()
		date := e.DOM.Find("div.m-post__meta span.date time").Text()
		dateNumber, _ := e.DOM.Find("div.m-post__meta span.date time").Attr("datetime")
		dateNumber = strings.TrimRight(dateNumber, "+00:00")
		content := e.DOM.Find("div.m-post-content").Text()

		c := session.DB("test").C("crawlerStackOF")
		if checkDate(dateNumber, &lastDate) {
			err := c.Insert(&blog{title, author, date, photo, content})
			if err != nil {
				log.Fatal(err)
			}
		}

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	for i := 1; i < 170; i++ {
		url := fmt.Sprint("https://stackoverflow.blog/company/page/", i)

		c.Visit(url)
	}

	errUpdate = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Date1"))
		if bucket == nil {
			panic("second update can not find Date1 !")
		}
		s := strconv.Itoa(lastDate)
		err := bucket.Put([]byte("lastDate"), []byte(s))
		return err
	})

	if errUpdate != nil {
		fmt.Println(errUpdate)
	}
}

func checkDate(dateNumber string, lastDate *int) bool {
	d := dateToInt(dateNumber)

	if acount == 1 {
		atemporary = d
	}
	if d > *lastDate {
		acount++
		return true
	} else {
		*lastDate = atemporary
		acount++
		return false
	}
}

func dateToInt(dateNumber string) int {
	ss := []string{"00", "00", "00"}

	a := strings.SplitN(dateNumber, "T", 2)
	sDay := strings.SplitN(a[0], "-", 3)
	sSecond := strings.SplitN(a[1], ":", 3)
	if len(sSecond) == 2 {
		ss[0] = sSecond[0]
		ss[1] = sSecond[1]
		sSecond = ss
	}
	if len(sSecond) == 1 {
		ss[0] = sSecond[0]
		sSecond = ss
	}
	if len(sSecond) == 0 {
		sSecond = ss
	}

	s := sDay[0] + sDay[1] + sDay[2] + sSecond[0] + sSecond[1] + sSecond[2]
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return i

}
