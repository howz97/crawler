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
	count0 int64 = 1
	temporary0 int  //暂时存放lastDate
)


func main(){
	var lastDate int

	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		var errAtoi error
		bucket1,err := tx.CreateBucket([]byte("Date"))
		if err != nil {
			lastDateSli := bucket1.Get([]byte("lastDate"))
			lastDate,errAtoi =strconv.Atoi(string(lastDateSli))
			return errAtoi
		}

		lastDate = 0
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
		content := e.DOM.Find("div.m-post-content").Text()

		if checkDate0(dateNumber, &lastDate){
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

	errUpdate :=db.Update(func(tx *bolt.Tx) error {
		bucket1 := tx.Bucket([]byte("Date"))
		s := strconv.Itoa(lastDate)
		err := bucket1.Put([]byte("lastDate"),[]byte(s))
		return err
	})

	if errUpdate != nil {
		fmt.Println(errUpdate)
	}
}

func checkDate0(dateNumber string,lastDate *int)bool{
	d := DateToInt0(dateNumber)
	if count0 == 1{
		temporary = d
	}
	if d > *lastDate {
		count0++
		return true
	}else {
		*lastDate = temporary0
		count0++
		return false
	}
}

func DateToInt0(dateNumber string)int{
	a := strings.Split(dateNumber,"T")
	s1 := strings.Split(a[0],"-")
	s2 := strings.Split(a[1],"-")
	s := s1[0]+s1[1]+s1[2]+s2[0]+s2[1]+s2[3]
	i,err := strconv.Atoi(s)
	if err != nil {
		panic("string to int error !")
	}

	return i

}

