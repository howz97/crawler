package segmentf

import (
	"fmt"
	"os"
	"time"

	"github.com/asciimoo/colly"
	"github.com/russross/blackfriday"

	"github.com/fengyfei/gu/libs/crawler"
	"github.com/fengyfei/gu/libs/logger"
)

type over struct{}

type segmentCrawler struct {
	collectorUrl  *colly.Collector
	collectorBlog *colly.Collector
	chUrl         chan string
	overClawler   chan over
}

// NewGoCNCrawler generates a crawler for segment blog.
func NewSegmentCrawler() crawler.Crawler {
	return &segmentCrawler{
		collectorUrl:  colly.NewCollector(),
		collectorBlog: colly.NewCollector(),
		chUrl:         make(chan string),
	}
}

// Crawler interface Init
func (c *segmentCrawler) Init() error {
	c.collectorUrl.OnHTML("article.ArticleInList.fade-in.ArticleInList--cardWhenBig", c.parse)
	c.collectorBlog.OnHTML("article.Article.fade-in.Article--featured", c.parseBlog)

	return nil
}

// Crawler interface Start
func (c *segmentCrawler) Start() error {
	go c.startUrl()
	for {
		select {
		case url := <-c.chUrl:
			err := c.startBlog(url)
			if err != nil {
				return err
			}
		case <-time.NewTimer(3 * time.Second).C:
			goto EXIT
		}
	}

EXIT:                   //是什么？？？？？？？？？？
	return nil
}

func (c *segmentCrawler) parse(e *colly.HTMLElement) {
	url, ok := e.DOM.Find("h2").Find("a").Attr("href")
	url = "https://segment.com" + url
	fmt.Println(url, ok)
	c.chUrl <- url
}

func (c *segmentCrawler) parseBlog(e *colly.HTMLElement) {
	title := e.DOM.Find("h2").Find("a").Text()
	html, _ := e.DOM.Html()                         //html()方法？？？？？？
	by := []byte(html)
	output := blackfriday.MarkdownBasic(by)          //是啥子？在干嘛？
	fmt.Println(string(output))

	f, _ := os.OpenFile("./blog/"+title+".md", os.O_CREATE|os.O_RDWR, 0644)     //打开本地文件？？？
	f.Write([]byte(output))
}

func (c *segmentCrawler) startUrl() {
	for i := 1; ; i++ {
		if i == 1 {
			c.collectorUrl.Visit("https://segment.com/blog/")
		} else {
			var url []string       //在这里定义变量？  这个string切片不会每循环重置？
			url = append(url, fmt.Sprint("https://segment.com/blog/page/", i))     //实现了翻页！！！
			err := c.collectorUrl.Visit(url[0])
			if err != nil {
				logger.Error("Visit Error:", err)
				break
			}
		}
	}
}

func (c *segmentCrawler) startBlog(url string) error {
	err := c.collectorBlog.Visit(url)
	if err != nil {
		logger.Error("Visit Error:", err)
		return err
	}
	return nil
}
