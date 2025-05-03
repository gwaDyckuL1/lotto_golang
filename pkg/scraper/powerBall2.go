/*
Currently this version is not being used.
It was created when I though I had to click a button.
There is an bug here, where if the first page is blank. It hangs up.
Maybe one day I'll come back and figure out a good fix.
*/

package scraper

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func ScrapingPowerBall2(db *sql.DB, w io.Writer) {
	createSQLTable(db)
	pbURL := creatingPBURL(db)
	//pbURL := `https://www.powerball.com/previous-results?ed=2025-04-26&gc=powerball&sd=2024-12-26`
	fmt.Println(pbURL)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	page := 1
	for {

		fullURL := fmt.Sprintf("%s&pg=%d", pbURL, page)

		err := chromedp.Run(ctx,
			chromedp.Navigate(fullURL),
			chromedp.WaitVisible("#searchNumbersResults"),
		)
		if err != nil {
			fmt.Println("Error navigating to Powerball website", err)
			return
		}
		fmt.Println("Powerball page", page, " loaded")

		var danger []*cdp.Node
		err = chromedp.Run(ctx,
			chromedp.Nodes(".card", &danger, chromedp.ByQueryAll),
		)
		if err != nil {
			fmt.Println("Error running last page check")
		}
		if len(danger) < 1 {
			return
		}
		fmt.Println(len(danger))

		var result []string
		chromedp.Run(ctx, chromedp.Evaluate(`Array.from(document.querySelectorAll(".card")).map(x => x.textContent)`, &result))

		var fixedResult [][]string
		for _, card := range result {
			space := regexp.MustCompile(`\s+`)
			fixed := []string{strings.TrimSpace(space.ReplaceAllString(card, " "))}
			fixedResult = append(fixedResult, fixed)
		}

		type Drawing struct {
			Date       string
			WhiteBall1 int
			WhiteBall2 int
			WhiteBall3 int
			WhiteBall4 int
			WhiteBall5 int
			PowerBall  int
		}

		for _, card := range fixedResult {
			for _, line := range card {
				var dateFound bool
				var eachDrawing Drawing
				var numStartIdx int
				for idx := range line {
					if !dateFound {
						date := line[:idx]
						parsedDate, err := time.Parse(`Mon, Jan 2, 2006`, date)
						if err == nil {
							dateFound = true
							eachDrawing.Date = parsedDate.Format("2006-01-02")
							numStartIdx = idx
						}
					}
					if line[idx] == 'P' {
						lottoNumbers := line[numStartIdx : idx-1]
						_, err := fmt.Sscanf(lottoNumbers, "%d %d %d %d %d %d",
							&eachDrawing.WhiteBall1, &eachDrawing.WhiteBall2, &eachDrawing.WhiteBall3, &eachDrawing.WhiteBall4, &eachDrawing.WhiteBall5, &eachDrawing.PowerBall)
						if err != nil {
							fmt.Println("Error parsing numbers into struct")
						}
						stmt, err := db.Prepare(`
						INSERT INTO powerBall (PlayDate, WhiteBall1, WhiteBall2, WhiteBall3, WhiteBall4, WhiteBall5, PowerBall)
						VALUES (?,?,?,?,?,?,?)
					`)
						if err != nil {
							log.Fatal("Unable to prepare powerBall table: ", err)
						}
						defer stmt.Close()

						_, err = stmt.Exec(
							eachDrawing.Date, eachDrawing.WhiteBall1, eachDrawing.WhiteBall2, eachDrawing.WhiteBall3, eachDrawing.WhiteBall4, eachDrawing.WhiteBall5, eachDrawing.PowerBall,
						)
						if err != nil {
							log.Println("Unable to insert draw date: ", eachDrawing.Date, err)
						} else {
							fmt.Println("Inserted draw data for date: ", eachDrawing.Date)
						}

						break
					}
				}
			}
		}
		if len(danger) < 30 {
			fmt.Println("No more pages")
			break
		}
		page++
	}
}

func creatingPBURL(db *sql.DB) string {
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
	return u.String()
}

// func createSQLTable(db *sql.DB) {
// 	createTableSQL := `
// 	CREATE TABLE IF NOT EXISTS powerBall (
// 	PlayDate STRING PRIMARY KEY,
// 	WhiteBall1 INTEGER,
// 	WhiteBall2 INTEGER,
// 	WhiteBall3 INTEGER,
// 	WhiteBall4 INTEGER,
// 	WhiteBall5 INTEGER,
// 	PowerBall INTEGER
// 	);`

// 	_, err := db.Exec(createTableSQL)
// 	if err != nil {
// 		log.Fatalf("Error creating table: %s", err)
// 	}
// }
