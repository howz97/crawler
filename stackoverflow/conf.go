package stackoverflow

import (
	"fmt"
	"github.com/spf13/viper"
)

var Conf conf

type conf struct {
	Site          string
	Database      string
	Collection    string
	DatabaseURI   string
	DSN           string
	UserAgentFile string
	MaxRetryTimes int
	Selector      sel
}

type sel struct {
	pSel pageSel
	aSel artclSel
}

type pageSel struct {
	articleURL string
	nextButton string
}

type artclSel struct {
	wholeArticle string
	author       string
	title        string
	date         string
	content      string
	tag          string
	comment      string
	cmntAuthor   string
	cmntDate     string
	cmntContent  string
}

func LoadConfig() {
	viper.SetConfigFile("/home/zhanghao/code/zh1014/crawler/stackoverflow/conf.json")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	Conf.Site = viper.GetString("site")
	Conf.Database = viper.GetString("database")
	Conf.Collection = viper.GetString("collection")
	Conf.DatabaseURI = viper.GetString("db_uri")
	Conf.DSN = viper.GetString("dsn")
	Conf.UserAgentFile = viper.GetString("user_agent_file")
	Conf.MaxRetryTimes = viper.GetInt("max_retry_times")

	Conf.Selector.pSel.articleURL = viper.GetString("selector.page.article_url")
	Conf.Selector.pSel.nextButton = viper.GetString("selector.page.next_button")

	Conf.Selector.aSel.wholeArticle = viper.GetString("selector.article.whole_article")
	Conf.Selector.aSel.author = viper.GetString("selector.article.author")
	Conf.Selector.aSel.title = viper.GetString("selector.article.title")
	Conf.Selector.aSel.date = viper.GetString("selector.article.date")
	Conf.Selector.aSel.content = viper.GetString("selector.article.content")
	Conf.Selector.aSel.tag = viper.GetString("selector.article.tag")
	Conf.Selector.aSel.comment = viper.GetString("selector.article.comment")
	Conf.Selector.aSel.cmntAuthor = viper.GetString("selector.article.comment_author")
	Conf.Selector.aSel.cmntDate = viper.GetString("selector.article.comment_date")
	Conf.Selector.aSel.cmntContent = viper.GetString("selector.article.comment_content")
}
