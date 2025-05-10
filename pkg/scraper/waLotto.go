package scraper

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gocolly/colly"
)

func ScrapingWALotto(db *sql.DB) string {
	s := ("We've entered scraping for WA Lotto")

	createSQLTable := `
	CREATE TABLE IF NOT EXISTS WA_Lotto (
	PlayDate STRING PRIMARY KEY,
	Ball1 INT,
	Ball2 INT,
	Ball3 INT,
	Ball4 INT,
	Ball5 INT,
	Ball6 INT
	);`

	_, err := db.Exec(createSQLTable)
	if err != nil {
		log.Fatalln("Error creating table", err)
	}

	var mostRecentSQLDate sql.NullString
	err = db.QueryRow("SELECT MAX(PlayDate) FROM WA_Lotto").Scan(&mostRecentSQLDate)
	if err != nil {
		log.Fatal("Can't pulled latest date in WA_Lotto", err)
	}

	var newestDate time.Time
	if mostRecentSQLDate.Valid {
		parseDate, err := time.Parse("2006-01-02", mostRecentSQLDate.String)
		if err != nil {
			log.Fatal("Could not parse most recent WA_Lotto date", err)
		} else {
			newestDate = parseDate
		}
	}

	//Always pulling last 180 days. I have not found a way to pull by date like Mega Millions or Powerball
	url := "https://walottery.com/WinningNumbers/PastDrawings.aspx?gamename=lotto&unittype=day&unitcount=180"

	c := colly.NewCollector(colly.AllowedDomains("walottery.com"))

	c.OnRequest(func(r *colly.Request) {
		s += fmt.Sprintln("Visiting: ", url)
	})

	c.OnHTML(".table-viewport-large", func(e *colly.HTMLElement) {
		ballList := [6]int{}

		rawDate := e.ChildText("h2")
		parseDate, err := time.Parse("Mon, Jan 2, 2006", rawDate)
		if err != nil {
			log.Fatal("Failed to parse date from WA lotto", err)
		}

		if newestDate == parseDate {
			s += fmt.Sprintln("Reached existing data. Stopping collection at date: ", newestDate)
			return
		}

		e.ForEach(".game-balls li", func(idx int, el *colly.HTMLElement) {
			num, err := strconv.Atoi(el.Text)
			if err != nil {
				log.Fatal("Failed to get numbers from WA lotto", err)
			} else {
				ballList[idx] = num
			}
		})

		stmt, err := db.Prepare(`
		INSERT OR IGNORE INTO WA_Lotto (PlayDate, Ball1, Ball2, Ball3, Ball4, Ball5, Ball6)
		VALUES (?,?,?,?,?,?,?)
		`)
		if err != nil {
			log.Fatal("Unable to prepare table for WA Lotto insert")
		}
		defer stmt.Close()

		_, err = stmt.Exec(
			parseDate.Format("2006-01-02"),
			ballList[0],
			ballList[1],
			ballList[2],
			ballList[3],
			ballList[4],
			ballList[5],
		)
		if err != nil {
			log.Fatal("Error entering data into Wa Lotto for date", parseDate)
		}
	})

	c.OnScraped(func(_ *colly.Response) {
		s = fmt.Sprintln("Finished scraping and processing data for WA Lotto")
	})

	err = c.Visit(url)
	if err != nil {
		log.Fatal("Error visiting Wa Lotto website", err)
	}
	return s
}
