package analysis

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
)

type Game struct {
	Name           string
	NumOfBalls     int
	SpecialBall    bool
	MaxWhiteBall   int
	MaxSpecialBall int
	CostPerPlay    float64
}

var Powerball = Game{Name: "PowerBall", NumOfBalls: 6, SpecialBall: true, MaxWhiteBall: 69, MaxSpecialBall: 26, CostPerPlay: 2}
var MegaMillions = Game{Name: "Mega_Millions", NumOfBalls: 6, SpecialBall: true, MaxWhiteBall: 70, MaxSpecialBall: 24, CostPerPlay: 5}
var WAlotto = Game{Name: "WA_Lotto", NumOfBalls: 6, SpecialBall: false, MaxWhiteBall: 49, CostPerPlay: 0.50}

var game = map[string]Game{
	"Powerball":     Powerball,
	"Mega Millions": MegaMillions,
	"WA Lotto":      WAlotto,
}

func makeBarChart[T int | float64](whiteBallMap, specialBallMap map[int]T, g Game, countOrProb string) string {

	complete := fmt.Sprintf("Completed making bar graph for %s by %s\n", g.Name, countOrProb)
	return complete
}

func CountBalls(gameName string, db *sql.DB) string {
	g := game[gameName]
	whiteBallMap, specialBallMap := getData(g, db)
	return makeBarChart(whiteBallMap, specialBallMap, g, "count")
}

func getData(g Game, db *sql.DB) (map[int]int, map[int]int) {
	whiteBallMap := make(map[int]int)
	specialBallMap := make(map[int]int)

	for num := range g.MaxWhiteBall {
		whiteBallMap[num] = 0
		if g.SpecialBall {
			if num <= g.MaxSpecialBall {
				specialBallMap[num] = 0
			}
		}
	}

	whiteBallCount := g.NumOfBalls
	if g.SpecialBall {
		whiteBallCount--
		getNumsByColumn(db, g.Name, "SpecialBall", specialBallMap)
	}

	for num := 0; num < whiteBallCount; num++ {
		column := "Ball" + strconv.Itoa(num+1)
		getNumsByColumn(db, g.Name, column, whiteBallMap)
	}

	return whiteBallMap, specialBallMap
}

func getNumsByColumn(db *sql.DB, table string, columnName string, theMap map[int]int) {
	query := fmt.Sprintf(`SELECT %s,
	COUNT(*)
	FROM %s
	GROUP BY %s 	
`, columnName, table, columnName)

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalln("Error querying the table", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var number, count int
		if err := rows.Scan(&number, &count); err != nil {
			log.Fatalln("Error scanning row: ", err)
		}
		theMap[number] += count
	}

}

func Probabilities(gameName string, db *sql.DB) string {
	s := ""

	return s
}

func MonteCarlo(gameName string, db *sql.DB) string {
	s := ""

	return s
}
