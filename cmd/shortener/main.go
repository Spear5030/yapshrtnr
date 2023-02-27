package main

import (
	"fmt"
	"log"

	"github.com/Spear5030/yapshrtnr/internal/app"
	"github.com/Spear5030/yapshrtnr/internal/config"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Println("Build version: ", buildVersion)
	fmt.Println("Build date: ", buildDate)
	fmt.Println("Build commit: ", buildCommit)

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
