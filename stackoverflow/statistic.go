package stackoverflow

import (
	"github.com/fatih/camelcase"
	"github.com/zh1014/algorithm/alphabet"
	tsto "github.com/zh1014/algorithm/trie-tree/tst-optimized"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"strconv"
)

const (
	statisticFile = "/home/zhanghao/code/zh1014/crawler/stackoverflow/statistic.txt"
)

var (
	Letter = alphabet.NewAlphabet("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
)

func Statistic() {
	LoadConfig()
	trie := tsto.NewTst2(Letter)
	mgoSes, err := mgo.Dial(Conf.DSN)
	if err != nil {
		panic(err)
	}
	c := mgoSes.DB(Conf.Database).C(Conf.Collection)
	iter := c.Find(bson.M{}).Iter()
	rslt := &article{}
	for iter.Next(rslt) {
		words := rslt.wordsOfContent()
		for _, word := range words {
			if checkCharset(trie.Alpb, word) {
				addToTrie(trie, word)
			}
		}
	}
	var out *os.File
	out, err = os.OpenFile(statisticFile, os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	keys := trie.Keys()
	for _, k := range keys {
		c := strconv.Itoa(queryTrie(trie, k))
		out.WriteString(k + " " + c + "\n")
	}
}

func (a *article) wordsOfContent() []string {
	words := make([]string, 0, 512)
	words = append(words, camelcase.Split(a.Title)...)
	words = append(words, camelcase.Split(a.Content)...)
	for _, c := range a.Comments {
		words = append(words, camelcase.Split(c.Content)...)
	}
	return words
}

func addToTrie(t *tsto.Tst2, k string) {
	v := t.Find(k)
	if v == nil {
		t.Insert(k, 1)
	} else {
		t.Insert(k, v.(int)+1)
	}
}

func queryTrie(t *tsto.Tst2, k string) int {
	v := t.Find(k)
	if v == nil {
		return 0
	}
	return v.(int)
}

func checkCharset(a alphabet.IAlphabet, s string) bool {
	for _, r := range s {
		if !a.Contains(r) {
			return false
		}
	}
	return true
}
