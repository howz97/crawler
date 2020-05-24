package stackoverflow

import "strings"

const (
	trimCutset = " \n\t"
)

type article struct {
	Title    string     `bson:"title"`
	Author   string     `bson:"author"`
	Date     string     `bson:"date"`
	Content  string     `bson:"content"`
	Tags     []string   `bson:"tags"`
	Comments []*comment `bson:"comments"`
}

func (a *article) trim(cutset string) {
	a.Title = strings.Trim(a.Title, cutset)
	a.Author = strings.Trim(a.Author, cutset)
	a.Date = strings.Trim(a.Date, cutset)
	a.Content = strings.Trim(a.Content, cutset)
	for i := range a.Tags {
		a.Tags[i] = strings.Trim(a.Tags[i], cutset)
	}
	for i := range a.Comments {
		a.Comments[i].trim(cutset)
	}
}

type comment struct {
	Author  string `bson:"comment_author"`
	Content string `bson:"comment_content"`
	Date    string `bson:"comment_date"`
}

func (c *comment) trim(cutset string) {
	c.Author = strings.Trim(c.Author, cutset)
	c.Date = strings.Trim(c.Date, cutset)
	c.Content = strings.Trim(c.Content, cutset)
}
