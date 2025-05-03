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

func ScrapeMegaMillions(db *sql.DB, w io.Writer) {

	type XMLResponse struct {
		Body string `xml:",chardata"`
	}

	type Drawing struct {
		PlayDate    string `json:"PlayDate"`
		N1          int    `json:"N1"`
		N2          int    `json:"N2"`
		N3          int    `json:"N3"`
		N4          int    `json:"N4"`
		N5          int    `json:"N5"`
		MegaBall    int    `json:"MBall"`
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

	fmt.Fprintln(w, "Visiting:", u)

	resp, err := http.Get(u.String())
	if err != nil {
		fmt.Fprintln(w, "Error: ", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(w, "Read error: ", err)
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
		INSERT INTO megaMillion (PlayDate, N1, N2, N3, N4, N5, MegaBall, Megaplier)
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
			draw.N1, draw.N2, draw.N3, draw.N4, draw.N5,
			draw.MegaBall, draw.Megaplier,
		)
		if err != nil {
			log.Printf("Insert failed for %s: %v", draw.PlayDate, err)
		} else {
			fmt.Fprintln(w, "MegaMillion - Inserted draw:", draw.PlayDate)
		}
	}
}

func createMegaTable(db *sql.DB) {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS megaMillion (
	PlayDate STRING NOT NULL PRIMARY KEY,
	N1 INTEGER,
	N2 INTEGER,
	N3 INTEGER,
	N4 INTEGER,
	N5 INTEGER,
	MegaBall INTEGER,
	Megaplier INTEGER
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating table: %s", err)
	}

}

func getMostCurrentDate(db *sql.DB) time.Time {
	var mostRecent sql.NullString

	err := db.QueryRow("SELECT MAX(PlayDate) FROM megaMillion").Scan(&mostRecent)
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
