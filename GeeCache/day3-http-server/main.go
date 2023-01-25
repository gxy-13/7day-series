package main

import (
	"fmt"
	"geecache/geecache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Kevin":  "123",
	"Lebron": "456",
	"James":  "789",
}

func main() {
	geecache.NewGroup("scores", geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}), 2<<10)

	addr := "localhost:9999"
	peers := geecache.NewHttPPool(addr)
	log.Println("geeCache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
