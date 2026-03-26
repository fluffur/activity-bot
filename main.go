package main

import (
	"log"
	"time"
)

func main() {
	now := time.Now()
	location, err := time.LoadLocation("Asia/Novosibirsk")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(now.In(time.UTC))
	log.Println(now.In(location))
	log.Println(now.In(location).After(now.In(time.UTC)))
}
