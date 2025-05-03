package scraper

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/gocolly/colly"
)

func ScrapingPowerBall(db *sql.DB, w io.Writer) {

	type Drawing struct {
		Date       string
		WhiteBall1 int
		WhiteBall2 int
		WhiteBall3 int
		WhiteBall4 int
		WhiteBall5 int
		PowerBall  int
	}
	var eachDraw []Drawing
	seenDates := make(map[string]bool)

	createSQLTable(db)
	pbURL, recentDate := creatingURL(db)

	c := colly.NewCollector(
		colly.AllowedDomains("www.powerball.com"),
	)

	pageNum := 1
	for {

		pgURL := pbURL + "&pg=" + strconv.Itoa(pageNum)
		fmt.Fprintln(w, "Visiting", pgURL)
		var pageHadCards bool
		c.OnHTML("div.card-body", func(e *colly.HTMLElement) {
			pageHadCards = true
		})

		c.OnHTML("div.card-body", func(e *colly.HTMLElement) {
			//fmt.Println(e.Text)
			dailyDraw := Drawing{}
			whiteBallList := [5]int{}

			rawDate := e.ChildText(".card-title")
			parsedDate, err := time.Parse("Mon, Jan 2, 2006", rawDate)
			if err != nil {
				log.Fatal("Failed to parse date", err)
			}
			dailyDraw.Date = parsedDate.Format("2006-01-02")

			e.ForEach(".white-balls", func(idx int, el *colly.HTMLElement) {
				num, err := strconv.Atoi(el.Text)
				if err != nil {
					fmt.Println("Failed to convert number", num)
				} else {
					whiteBallList[idx] = num
				}
			})

			dailyDraw.WhiteBall1 = whiteBallList[0]
			dailyDraw.WhiteBall2 = whiteBallList[1]
			dailyDraw.WhiteBall3 = whiteBallList[2]
			dailyDraw.WhiteBall4 = whiteBallList[3]
			dailyDraw.WhiteBall5 = whiteBallList[4]

			powerBall, err := strconv.Atoi(e.ChildText(".powerball"))
			if err != nil {
				fmt.Println("Powerball number failed to convert", err)
			}
			dailyDraw.PowerBall = powerBall

			if !seenDates[dailyDraw.Date] {
				eachDraw = append(eachDraw, dailyDraw)
				seenDates[dailyDraw.Date] = true
			}

		})

		err := c.Visit(pgURL)
		if err != nil {
			log.Println("Visit error", err)
			break
		}

		if !pageHadCards {
			date, err := time.Parse("2006-01-02", recentDate)
			if err != nil {
				break
			} else {
				fmt.Fprintln(w, "Scraped Powerball information going back to", date.Format("2006-01-02"))
				break
			}
		}

		pageNum++
	}

	c.OnScraped(func(_ *colly.Response) {
		fmt.Fprintln(w, "Finished with powerball")
	})

	stmt, err := db.Prepare(`
		INSERT INTO powerBall (PlayDate, WhiteBall1, WhiteBall2, WhiteBall3, WhiteBall4, WhiteBall5, PowerBall)
		VALUES (?,?,?,?,?,?,?)
	`)
	if err != nil {
		log.Fatal("Unable to prepare powerBall table: ", err)
	}
	defer stmt.Close()

	for _, draw := range eachDraw {
		_, err := stmt.Exec(
			draw.Date, draw.WhiteBall1, draw.WhiteBall2, draw.WhiteBall3, draw.WhiteBall4, draw.WhiteBall5, draw.PowerBall,
		)
		if err != nil {
			log.Println("Unable to insert draw date: ", draw.Date, err)
		} else {
			fmt.Fprintln(w, "Inserted draw data for date: ", draw.Date)
		}
	}
}

func creatingURL(db *sql.DB) (string, string) {
	u, _ := url.Parse("https://www.powerball.com/previous-results?gc=powerball&sd=2024-02-14&ed=2024-04-13")
	var mostRecent sql.NullString

	err := db.QueryRow("SELECT MAX(PlayDate) FROM powerBall").Scan(&mostRecent)
	if err != nil {
		log.Fatal(err)
	}
	var dateStart time.Time
	if mostRecent.Valid {
		parsedTime, err := time.Parse("2006-01-02", mostRecent.String)
		if err != nil {
			fmt.Println("Failed to parse date")
		}
		dateStart = parsedTime.AddDate(0, 0, 1)
	} else {
		dateStart = time.Now().AddDate(-1, 0, 0)
	}

	query := url.Values{}
	query.Set("gc", "powerball")
	query.Set("sd", dateStart.Format("2006-01-02"))
	query.Set("ed", time.Now().Format("2006-01-02"))

	u.RawQuery = query.Encode()
	return u.String(), mostRecent.String
}

func createSQLTable(db *sql.DB) {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS powerBall (
	PlayDate STRING PRIMARY KEY,
	WhiteBall1 INTEGER,
	WhiteBall2 INTEGER,
	WhiteBall3 INTEGER,
	WhiteBall4 INTEGER,
	WhiteBall5 INTEGER,
	PowerBall INTEGER
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating table: %s", err)
	}
}
