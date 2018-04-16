package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	res, err := http.Get("https://stackoverflow.blog/wp-content/uploads/2017/02/jVjrp.png")
	if err != nil {
		panic(err)
	}
	file, err := os.Create("1.jpg")
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(file, res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println("下载完成！")
}
