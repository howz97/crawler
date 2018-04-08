package main

import (
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

func main() {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	errUpdate := db.Update(func(tx *bolt.Tx) error {
		bucket1 := tx.Bucket([]byte("first_bucket"))

		errPut := bucket1.Put([]byte("answer"),[]byte("3"))
		if errPut != nil{
			return  errPut
		}

		get1 := bucket1.Get([]byte("answer"))

		fmt.Println(string(get1))
		return nil
	})

	if errUpdate != nil{
		fmt.Println(errUpdate)
	}

	db.View(func(tx *bolt.Tx) error {
		bucket1 := tx.Bucket([]byte("first_bucket"))
		get1 := bucket1.Get([]byte("answer"))
		fmt.Println(string(get1))
		return nil
	})


}