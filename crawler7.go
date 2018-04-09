package main

import (
	"fmt"
	"strings"
	"strconv"
	"log"
	"time"

	"github.com/gocolly/colly"
	"github.com/boltdb/bolt"
)

// count is used to count how many times checkDate() was called .
var (
	count int64 = 1
	temporary int  //暂时存放lastDate
)


func main(){
	var lastDate int

	db, errOpen := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if errOpen != nil {
		log.Fatal(errOpen)
	}
	defer db.Close()

	errUpdate := db.Update(func(tx *bolt.Tx) error {
		bucket1, err := tx.CreateBucketIfNotExists([]byte("Date1"))
		if err != nil {
			log.Fatal("create bucket error ")
			return err
		}

		lastDateSli := bucket1.Get([]byte("lastDate"))
		if lastDateSli == nil {
			lastDate = 0
		}else {
			lastDate, err = strconv.Atoi(string(lastDateSli))
			if err != nil {
				log.Fatal(" Atoi error")
				return err
			}
		}

		return nil
	})

	c := colly.NewCollector()
	detailCollector := c.Clone()


	c.OnHTML("article div.m-post-card__content-column", func(e *colly.HTMLElement){

		link ,_ := e.DOM.Find("h2 a").Attr("href")
		detailCollector.Visit(link)

	})

	detailCollector.OnHTML("main#main.site-main", func(e *colly.HTMLElement) {

		title := e.DOM.Find("div.column h1").Text()
		title = strings.TrimSpace(title)
		photo,_ := e.DOM.Find("div span span a img").Attr("src")
		author := e.DOM.Find("div.m-post__meta span.author-name a").Text()
		date := e.DOM.Find("div.m-post__meta span.date time").Text()
		dateNumber,_ := e.DOM.Find("div.m-post__meta span.date time").Attr("datetime")
		dateNumber = strings.TrimRight(dateNumber,"+00:00")
		content := e.DOM.Find("div.m-post-content").Text()

		if checkDate(dateNumber, &lastDate){
			fmt.Printf("<<%s>> \n%s     on %s \nphoto :%s \n%s\n----------------------------------------------------\n",
				title,author,date,photo,content)
		}

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting",r.URL.String())
	})

	for i :=1;i<170 ;i++  {
		url := fmt.Sprint("https://stackoverflow.blog/company/page/",i)

		c.Visit(url)
	}

	errUpdate =db.Update(func(tx *bolt.Tx) error {
		bucket1 := tx.Bucket([]byte("Date1"))
		if bucket1 == nil {
			panic("second update can not find Date1 !!!!!!!!")
		}
		s := strconv.Itoa(lastDate)
		err := bucket1.Put([]byte("lastDate"),[]byte(s))
		return err
	})

	if errUpdate != nil {
		fmt.Println(errUpdate)
	}
}

func checkDate(dateNumber string,lastDate *int)bool{
	d := DateToInt(dateNumber)

	if count == 1{
		temporary = d
	}
	if d > *lastDate {
		count++
		return true
	}else {
		*lastDate = temporary
		count++
		return false
	}
}

func DateToInt(dateNumber string)int{
	ss := []string{"00","00","00"}

	a := strings.SplitN(dateNumber,"T",2)
	s1 := strings.SplitN(a[0],"-",3)
	s2 := strings.SplitN(a[1],":",3)
	if len(s2) == 2 {
		ss[0] = s2[0]
		ss[1] = s2[1]
		s2 = ss
	}
	if len(s2) == 1 {
		ss[0] =s2[0]
		s2 = ss
	}
	if len(s2) == 0 {
		s2 = ss
	}

    s := s1[0]+s1[1]+s1[2]+s2[0]+s2[1]+s2[2]
    i,err := strconv.Atoi(s)
	if err != nil {
		panic("string to int error !")
	}

	return i

}
