package main

import (
	"fmt"
	"strings"
	"strconv"
	"github.com/gocolly/colly"

)

// count is used to count how many times checkDate() was called .
var (
	count0 int64 = 1
	temporary0 int  //暂时存放lastDate
)


func main(){
//	var lastDate = 20180327113012
	var lastDate = 20180000000000

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
		fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~",dateNumber)
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
}

func checkDate0(dateNumber string,lastDate0 *int)bool{
	d := DateToInt0(dateNumber)
	if count0 == 1{
		temporary0 = d
	}
	if d > *lastDate0 {
		count0++
		return true
	}else {
		*lastDate0 = temporary0
		count0++
		return false
	}
}

func DateToInt0(dateNumber string)int{
	fmt.Println("111111111111111")

	a := strings.Split(dateNumber,"T")
	fmt.Println("22222222222222222")

	s1 := strings.Split(a[0],"-")
	fmt.Println("33333333333333333")
	fmt.Println(s1)
	z := s1[0]+s1[1]+s1[2]
	fmt.Println("年月日：",z)

	s2 := strings.Split(a[1],":")
	fmt.Println("44444444444444444")
	fmt.Println(s2)
	h := s2[0]+s2[1]+s2[2]
	fmt.Println("时分秒：",h)


//	s := s1[0]+s1[1]+s1[2]+s2[0]+s2[1]+s2[2]
    s := z + h
	fmt.Println("string s:",s)
	i,err := strconv.Atoi(s)
	fmt.Println("int i : ",i)

	if err != nil {
		panic("string to int error !")
	}

	return i

}
