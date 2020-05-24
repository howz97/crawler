package stackoverflow

import (
	"fmt"
	"testing"
)

func TestStackOverFlow(t *testing.T) {
	LoadConfig()
	crwlr := New()
	crwlr.Init()
	crwlr.Run()
}

func TestFetchArticle(t *testing.T) {
	crwlr := New()
	crwlr.Init()
	crwlr.visitArticle("https://stackoverflow.blog/2020/01/21/scripting-the-future-of-stack-2020-plans-vision/")
}

func TestTest(t *testing.T) {
	s := "/a大bc/"
	s = s[1:3]
	fmt.Println(s)
}
