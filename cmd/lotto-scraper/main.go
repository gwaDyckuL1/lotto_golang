package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	fmt.Print("\033[H\033[2J")
	db, err := sql.Open("sqlite", "data/lotto.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mainTeaTerminal(db)

	//scraper.ScrapeMegaMillions(db)
	//scraper.ScrapingPowerBall(db)
	//scraper.ScrapingPowerBall2(db)
	//analysis.AnalyzeMegaMillions(db)
}
