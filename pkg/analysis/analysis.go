package analysis

import (
	"database/sql"
	"fmt"
	"image/color"
	"log"
	"strconv"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
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
	p := plot.New()
	p.Title.Text = fmt.Sprintf("%s\nBall %s", g.Name, countOrProb)
	p.Y.Label.Text = countOrProb

	whiteValues := make(plotter.Values, 0, len(whiteBallMap))
	specialValues := make(plotter.Values, 0, len(whiteBallMap))
	xLabels := make([]string, 0, len(whiteBallMap))

	for i := 1; i <= g.MaxWhiteBall; i++ {
		whiteValues = append(whiteValues, float64(whiteBallMap[i]))
		specialValues = append(specialValues, float64(specialBallMap[i]))
		xLabels = append(xLabels, fmt.Sprintf("%d", i))
	}

	wBar, err := plotter.NewBarChart(whiteValues, vg.Points(5))
	if err != nil {
		return fmt.Sprintln("Ran into a problem creating wBar", err)
	}
	wBar.LineStyle.Width = vg.Length(0)
	wBar.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}

	sBar, err := plotter.NewBarChart(specialValues, vg.Points(5))
	if err != nil {
		return fmt.Sprintln("Ran into a problem creating sBar", err)
	}
	sBar.LineStyle.Width = vg.Length(0)
	sBar.Color = color.RGBA{R: 255, G: 215, B: 0, A: 255}

	wBar.Offset = -vg.Points(2.5)
	sBar.Offset = vg.Points(2.5)

	p.Add(wBar, sBar)
	p.NominalX(xLabels...)

	tableName := "graphs/" + g.Name + countOrProb + ".png"
	width := vg.Length(len(whiteBallMap)) * vg.Points(15)
	height := vg.Inch * 6

	p.Save(width, height, tableName)

	complete := fmt.Sprintf("Completed making bar graph for %s by %s\n\nLook in the graphs folder.", g.Name, countOrProb)
	return complete
}

func CountBalls(gameName string, db *sql.DB) string {
	g := game[gameName]
	whiteBallMap, specialBallMap := getData(g, db)
	return makeBarChart(whiteBallMap, specialBallMap, g, "Count")
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
