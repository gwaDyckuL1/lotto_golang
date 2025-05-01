package main

import (
	"database/sql"
	"log"

	"github.com/gwaDyckuL1/lotto-scraper/pkg/analysis"
	_ "modernc.org/sqlite"
)

func main() {

	db, err := sql.Open("sqlite", "data/lotto.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//scraper.ScrapeMegaMillions(db)
	//scraper.ScrapingPowerBall(db)
	//scraper.ScrapingPowerBall2(db)
	analysis.AnalyzeMegaMillions(db)
}
