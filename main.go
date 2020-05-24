package main

import "github.com/zh1014/crawler/stackoverflow"

func main() {
	stackoverflow.LoadConfig()
	crwlr := stackoverflow.New()
	crwlr.Init()
	crwlr.Run()
}
