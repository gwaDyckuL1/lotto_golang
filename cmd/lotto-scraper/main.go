package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gwaDyckuL1/lotto-scraper/ui"
	_ "modernc.org/sqlite"
)

func main() {

	requiredDirs := []string{"data", "graphs"}
	for _, dir := range requiredDirs {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create folder %s: %v", dir, err)
		}
	}

	fmt.Print("\033[H\033[2J")
	db, err := sql.Open("sqlite", "data/lotto.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ui.TeaTerminal(db)
}
