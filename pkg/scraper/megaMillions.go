package scraper

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"time"

	"database/sql"
	"io"
	"net/http"
	"net/url"
)

func ScrapeMegaMillions(db *sql.DB) string {
	s := ""
	type XMLResponse struct {
		Body string `xml:",chardata"`
	}

	type Drawing struct {
		PlayDate    string `json:"PlayDate"`
		Ball1       int    `json:"N1"`
		Ball2       int    `json:"N2"`
		Ball3       int    `json:"N3"`
		Ball4       int    `json:"N4"`
		Ball5       int    `json:"N5"`
		SpecialBall int    `json:"MBall"`
		Megaplier   int    `json:"Megaplier"`
		UpdatedBy   string `json:"UpdatedBy"`
		UpdatedTime string `json:"UpdatedTime"`
	}

	type MegaMillionsData struct {
		DrawingData []Drawing `json:"DrawingData"`
	}

	createMegaTable((db))
	recentDate := getMostCurrentDate((db))

	u, _ := url.Parse("https://www.megamillions.com/cmspages/utilservice.asmx/GetDrawingPagingData")

	query := url.Values{}
	query.Set("pageNumber", "1")
	query.Set("pageSize", "200")
	query.Set("startDate", recentDate.Format("01/02/2006"))
	query.Set("endDate", time.Now().Format("01/02/2006"))

	u.RawQuery = query.Encode()

	s += fmt.Sprintln("Visited:", u)

	resp, err := http.Get(u.String())
	if err != nil {
		problem := fmt.Sprintln("Error: ", err)
		return problem
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Read error: ", err)
	}

	var xmlResp XMLResponse
	if err := xml.Unmarshal(bodyBytes, &xmlResp); err != nil {
		panic(err)
	}

	var results MegaMillionsData
	if err := json.Unmarshal([]byte(xmlResp.Body), &results); err != nil {
		panic(err)
	}

	stmt, err := db.Prepare(`
		INSERT INTO Mega_Millions (PlayDate, Ball1, Ball2, Ball3, Ball4, Ball5, SpecialBall, Megaplier)
		VALUES (?,?,?,?,?,?,?,?)
	`)
	if err != nil {
		log.Fatal("Error preparing statement for megaMillion:", err)
	}
	defer stmt.Close()

	for _, draw := range results.DrawingData {
		parsedTime, oops := time.Parse("2006-01-02T15:04:05", draw.PlayDate)
		if oops != nil {
			log.Fatal("Error parsing date", err)
		}
		_, err = stmt.Exec(
			parsedTime.Format("2006-01-02"),
			draw.Ball1, draw.Ball2, draw.Ball3, draw.Ball4, draw.Ball5,
			draw.SpecialBall, draw.Megaplier,
		)
		if err != nil {
			log.Printf("Insert failed for %s: %v", draw.PlayDate, err)
		}
	}
	s += fmt.Sprintln("Scraped Mega Million information going back to", recentDate.Format("2006-01-02"))
	return s
}

func createMegaTable(db *sql.DB) {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS Mega_Millions (
	PlayDate STRING NOT NULL PRIMARY KEY,
	Ball1 INTEGER,
	Ball2 INTEGER,
	Ball3 INTEGER,
	Ball4 INTEGER,
	Ball5 INTEGER,
	SpecialBall INTEGER,
	Megaplier INTEGER
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating table: %s", err)
	}

}

func getMostCurrentDate(db *sql.DB) time.Time {
	var mostRecent sql.NullString

	err := db.QueryRow("SELECT MAX(PlayDate) FROM Mega_Millions").Scan(&mostRecent)
	if err != nil {
		log.Fatal(err)
	}
	if mostRecent.Valid {
		parsedTime, err := time.Parse("2006-01-02", mostRecent.String)
		if err != nil {
			log.Println("Error parsing date string from database:", err)
			return time.Now().AddDate(-1, 0, 0) // fallback
		}
		return parsedTime.AddDate(0, 0, 1) // move forward 1 day
	}
	return time.Now().AddDate(-1, 0, 0) // fallback to 1 year ago
}
