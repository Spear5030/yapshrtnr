package main

import (
	"github.com/Spear5030/yapshrtnr/internal/app"
	"github.com/Spear5030/yapshrtnr/internal/config"
	"log"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(a.Run())
}
