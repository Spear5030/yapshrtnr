package main

import (
	"github.com/Spear5030/yapshrtnr/internal/app"
	"log"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(a.Run())
}
