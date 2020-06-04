// Copyright 2018 The zh1014. All rights reserved.
package stackoverflow

import (
	"bufio"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
	collyMgo "github.com/zolamk/colly-mongo-storage/colly/mongo"
	"gopkg.in/mgo.v2"
	"os"
	"time"
)

const (
	ctxKeyRetryTimes = "RetryTimes"
)

type Crawler struct {
	userAgents       []string
	pageCollector    *colly.Collector
	articleCollector *colly.Collector
	mgoS             *mgo.Session
}

func New() *Crawler {
	crwlr := &Crawler{
		pageCollector:    colly.NewCollector(),
		articleCollector: colly.NewCollector(),
	}
	return crwlr
}

func (crwlr *Crawler) Init() {
	logrus.SetLevel(logrus.DebugLevel)
	var err error
	crwlr.mgoS, err = mgo.Dial(Conf.DSN)
	if err != nil {
		panic(err)
	}
	crwlr.loadUserAgents(Conf.UserAgentFile)
	crwlr.initPageCllt()
	crwlr.initArticleCllt()
}

func (crwlr *Crawler) initPageCllt() {
	crwlr.pageCollector.OnRequest(func(r *colly.Request) {
		logrus.Infof("requesting page %v", r.URL.String())
	})
	crwlr.pageCollector.OnHTML(Conf.Selector.pSel.articleURL, crwlr.onArticleLink)
	crwlr.pageCollector.OnHTML(Conf.Selector.pSel.nextButton, crwlr.onNextButton)
	crwlr.pageCollector.OnError(errCallback)
	crwlr.pageCollector.SetRequestTimeout(30 * time.Second)
}

func (crwlr *Crawler) initArticleCllt() {
	crwlr.articleCollector.OnRequest(func(r *colly.Request) {
		logrus.Infof("requesting article %v", r.URL.String())
	})
	crwlr.articleCollector.OnHTML(Conf.Selector.aSel.wholeArticle, crwlr.onArticle)

	storage := &collyMgo.Storage{
		Database: Conf.Database,
		URI:      Conf.DatabaseURI,
	}
	if err := crwlr.articleCollector.SetStorage(storage); err != nil {
		panic(err)
	}
	crwlr.pageCollector.SetRequestTimeout(30 * time.Second)
}

func (crwlr *Crawler) loadUserAgents(filename string) {
	f, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		crwlr.userAgents = append(crwlr.userAgents, scanner.Text())
	}
	crwlr.articleCollector.OnError(errCallback)
}

func errCallback(r *colly.Response, err error) {
	logrus.Errorf("%v\nUser-Agent:%v", err.Error(), r.Request.Headers.Get("User-Agent"))
	itf := r.Ctx.GetAny(ctxKeyRetryTimes)
	rt := 0
	if itf != nil {
		rt = itf.(int)
	}
	if rt < Conf.MaxRetryTimes {
		rt++
		logrus.Infof("retrying for the %v time", rt)
		r.Ctx.Put(ctxKeyRetryTimes, rt)
		r.Request.Retry()
	}
}

func (crwlr *Crawler) Run() {
	crwlr.visitPage(Conf.Site)
	crwlr.pageCollector.Wait()
	crwlr.articleCollector.Wait()
}

func (crwlr *Crawler) visitPage(url string) {
	crwlr.pageCollector.UserAgent = crwlr.chooseUserAgent()
	crwlr.pageCollector.Visit(url)
}

func (crwlr *Crawler) visitArticle(url string) {
	crwlr.articleCollector.UserAgent = crwlr.chooseUserAgent()
	crwlr.articleCollector.Visit(url)
}

func (crwlr *Crawler) chooseUserAgent() string {
	unixNano := int(time.Now().UnixNano())
	return crwlr.userAgents[unixNano%len(crwlr.userAgents)]
}

func (crwlr *Crawler) onArticleLink(e *colly.HTMLElement) {
	link, exists := e.DOM.Attr("href")
	if !exists {
		logrus.Error("no link in this HTML element")
		return
	}
	go crwlr.visitArticle(link)
	//crwlr.visitArticle(link)
}

func (crwlr *Crawler) onNextButton(e *colly.HTMLElement) {
	link, exists := e.DOM.Attr("href")
	if !exists {
		logrus.Error("no link in this HTML element")
		return
	}
	crwlr.visitPage(link)
}

func (crwlr *Crawler) onArticle(e *colly.HTMLElement) {
	atc := &article{}
	atc.Author = e.DOM.Find(Conf.Selector.aSel.author).Text()
	atc.Title = e.DOM.Find(Conf.Selector.aSel.title).Text()
	atc.Date = e.DOM.Find(Conf.Selector.aSel.date).Text()
	atc.Content = e.DOM.Find(Conf.Selector.aSel.content).Text()
	// tags
	e.DOM.Find(Conf.Selector.aSel.tag).Each(func(_ int, sel *goquery.Selection) {
		atc.Tags = append(atc.Tags, sel.Text())
	})
	// comments
	e.DOM.Find(Conf.Selector.aSel.comment).Each(func(_ int, sel *goquery.Selection) {
		cmnt := &comment{}
		cmnt.Author = sel.Find(Conf.Selector.aSel.cmntAuthor).Text()
		cmnt.Date = sel.Find(Conf.Selector.aSel.cmntDate).Text()
		cmnt.Content = sel.Find(Conf.Selector.aSel.cmntContent).Text()
		atc.Comments = append(atc.Comments, cmnt)
	})
	atc.trim(trimCutset)
	err := crwlr.mgoS.DB(Conf.Database).C(Conf.Collection).Insert(atc)
	if err != nil {
		logrus.Errorf("failed to save <<%v>>: %v", atc.Title, err.Error())
	}
}
