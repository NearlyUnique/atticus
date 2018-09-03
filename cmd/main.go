package main

import (
	"log"

	"github.com/NearlyUnique/atticus"
)

func main() {
	err := atticus.New().Run()

	if err != nil {
		log.Printf("terminated:%v", err)
	}
}
