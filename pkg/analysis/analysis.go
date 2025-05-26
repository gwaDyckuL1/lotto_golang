package analysis

import (
	"database/sql"
	"fmt"
	"image/color"
	"log"
	"math/rand/v2"
	"slices"
	"sort"
	"strconv"

	"github.com/dustin/go-humanize"
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
	WaysToWin      map[string]int
}

var Powerball = Game{Name: "Powerball", NumOfBalls: 6, SpecialBall: true, MaxWhiteBall: 69, MaxSpecialBall: 26, CostPerPlay: 2, WaysToWin: map[string]int{
	"0:true":  4,
	"1:true":  4,
	"2:true":  7,
	"3:false": 7,
	"3:true":  100,
	"4:false": 100,
	"4:true":  50000,
	"5:false": 1000000,
	"5:true":  0,
}}
var MegaMillions = Game{Name: "Mega_Millions", NumOfBalls: 6, SpecialBall: true, MaxWhiteBall: 70, MaxSpecialBall: 24, CostPerPlay: 5, WaysToWin: map[string]int{
	"0:true":  10,
	"1:true":  14,
	"2:true":  20,
	"3:false": 20,
	"3:true":  400,
	"4:false": 1000,
	"4:true":  20000,
	"5:false": 2000000,
	"5:true":  0,
}}
var WAlotto = Game{Name: "'WA Lotto'", NumOfBalls: 6, SpecialBall: false, MaxWhiteBall: 49, CostPerPlay: 0.50, WaysToWin: map[string]int{
	"3:false": 3,
	"4:false": 30,
	"5:false": 1000,
	"6:false": 0,
}}

var game = map[string]Game{
	"Powerball":     Powerball,
	"Mega_Millions": MegaMillions,
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

	p.Legend.Add("White Balls", wBar)
	if g.SpecialBall {
		p.Legend.Add("Sepcial Ball", sBar)
	}
	p.Legend.Top = true
	p.Legend.Left = true

	tableName := "graphs/" + g.Name + countOrProb + ".png"
	width := vg.Length(len(whiteBallMap)) * vg.Points(15)
	height := vg.Inch * 6

	p.Save(width, height, tableName)

	complete := fmt.Sprintf("Completed making bar graph for %s by %s\n\nLook in the graphs folder.", g.Name, countOrProb)
	return complete
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

func Probabilities(whiteBallMap, specialBallMap map[int]int, g Game, db *sql.DB) (map[int]float64, map[int]float64) {

	whiteTotal, specialTotal := 0, 0

	for _, val := range whiteBallMap {
		whiteTotal += val
	}
	for _, val := range specialBallMap {
		specialTotal += val
	}
	whiteBallProb := make(map[int]float64)
	specialBallProb := make(map[int]float64)

	for i := 1; i <= g.MaxWhiteBall; i++ {
		whiteBallProb[i] = float64(whiteBallMap[i]) / float64(whiteTotal) * 100.00
		if i <= g.MaxSpecialBall {
			specialBallProb[i] = float64(specialBallMap[i]) / float64(specialTotal) * 100.00
		}
	}
	return whiteBallProb, specialBallProb
}

func MonteCarlo(gameName string, db *sql.DB) string {
	s := "Monty predicts you should pick...\n\n"
	g := game[gameName]

	whiteBallsCountMap, specialBallsCountMap := getData(g, db)

	totalWhiteBalls, totalSpecialBalls := 0, 0

	for _, val := range whiteBallsCountMap {
		totalWhiteBalls += val
	}
	if g.SpecialBall {
		for _, val := range specialBallsCountMap {
			totalSpecialBalls += val
		}
	}

	scaleFactor := 1000
	weightedWhite := []int{}
	weightedSpecial := []int{}

	for i := 1; i <= g.MaxWhiteBall; i++ {
		probofNum := float64(whiteBallsCountMap[i]) / float64(totalWhiteBalls)
		howManyWhiteToAdd := int(probofNum * float64(scaleFactor))
		for range howManyWhiteToAdd {
			weightedWhite = append(weightedWhite, i)
		}
		if g.SpecialBall && i <= g.MaxSpecialBall {
			probofSpecial := float64(specialBallsCountMap[i]) / float64(totalSpecialBalls)
			howManySpeicalToAdd := int(probofSpecial * float64(scaleFactor))
			for range howManySpeicalToAdd {
				weightedSpecial = append(weightedSpecial, i)
			}
		}
	}

	montyWhiteBalls := make(map[int]int)
	montySpecialBalls := make(map[int]int)

	for i := 1; i <= g.MaxWhiteBall; i++ {
		montyWhiteBalls[i] = 0
		if g.SpecialBall && i <= g.MaxSpecialBall {
			montySpecialBalls[i] = 0
		}
	}

	if g.SpecialBall {
		g.NumOfBalls--
	}

	for i := 0; i < 10000; i++ {
		rand.Shuffle(len(weightedWhite), func(i, j int) {
			weightedWhite[i], weightedWhite[j] = weightedWhite[j], weightedWhite[i]
		})
		if g.SpecialBall {
			rand.Shuffle(len(weightedSpecial), func(i, j int) {
				weightedSpecial[i], weightedSpecial[j] = weightedSpecial[j], weightedSpecial[i]
			})
			montySpecialBalls[weightedSpecial[0]]++
		}

		seen := map[int]bool{}
		ballCount := 0

		for _, val := range weightedWhite {
			if !seen[val] {
				seen[val] = true
				montyWhiteBalls[val]++
				ballCount++
			}
			if ballCount == g.NumOfBalls {
				break
			}
		}
	}

	sortedWhiteCount := [][]int{}
	for key, val := range montyWhiteBalls {
		sortedWhiteCount = append(sortedWhiteCount, []int{key, val})
	}

	sort.Slice(sortedWhiteCount, func(i, j int) bool {
		return sortedWhiteCount[i][1] > sortedWhiteCount[j][1]
	})

	s += "White Balls to select: "
	for i := 0; i < g.NumOfBalls; i++ {
		s += strconv.Itoa(sortedWhiteCount[i][0]) + " "
	}

	if g.SpecialBall {
		sortedSpecialCount := [][]int{}
		for key, val := range montySpecialBalls {
			sortedSpecialCount = append(sortedSpecialCount, []int{key, val})
		}

		sort.Slice(sortedSpecialCount, func(i, j int) bool {
			return sortedSpecialCount[i][1] > sortedSpecialCount[j][1]
		})

		s += "\nSpecial Ball to select: " + strconv.Itoa(sortedSpecialCount[0][0])
	}
	s += "\n\n"
	s += makeBarChart(montyWhiteBalls, montySpecialBalls, g, "Monty Count")
	return s
}

func MontysCostToWin(gameName string, db *sql.DB) string {
	g := game[gameName]
	s := fmt.Sprintf("Monty tried to win %s\n\n", g.Name)

	whiteBallCountMap, specialBallCountMap := getData(g, db)

	totalWhiteBalls, totalSpecialBalls := 0, 0

	for key := 1; key <= g.MaxWhiteBall; key++ {
		totalWhiteBalls += whiteBallCountMap[key]
	}

	if g.SpecialBall {
		for key := 1; key <= g.MaxSpecialBall; key++ {
			totalSpecialBalls += specialBallCountMap[key]
		}
	}

	weightedWhiteList := []int{}
	weightSpecialList := []int{}
	scaleFactor := 1000.00

	for key, val := range whiteBallCountMap {
		numProbability := float64(val) / float64(totalWhiteBalls)
		numOfKey := numProbability * scaleFactor
		for i := 0; i < int(numOfKey); i++ {
			weightedWhiteList = append(weightedWhiteList, key)
		}
	}
	if g.SpecialBall {
		for key, val := range specialBallCountMap {
			numProbability := float64(val) / float64(totalSpecialBalls)
			numOfKey := numProbability * scaleFactor
			for i := 0; i < int(numOfKey); i++ {
				weightSpecialList = append(weightSpecialList, key)
			}
		}
	}

	query := "SELECT PlayDate,"

	if g.SpecialBall {
		for i := 1; i < g.NumOfBalls; i++ {
			query += " Ball" + fmt.Sprintf("%d", i) + ","
		}
		query += " SpecialBall"
	} else {
		for i := 1; i < g.NumOfBalls; i++ {
			query += " Ball" + fmt.Sprintf("%d", i) + ","
		}
		query += " Ball" + fmt.Sprintf("%d", g.NumOfBalls)
	}
	query += fmt.Sprintf(`
		FROM %s
		WHERE PlayDate = (
			SELECT MAX(PlayDate)
			FROM %s
		)
	`, g.Name, g.Name)

	row := db.QueryRow(query)
	recentDraw := make([]int, g.NumOfBalls)
	values := make([]interface{}, g.NumOfBalls+1)
	var date string

	values[0] = &date

	for i := 1; i <= g.NumOfBalls; i++ {
		values[i] = &recentDraw[i-1]
	}

	err := row.Scan(values...)
	if err != nil {
		log.Fatal("Error scanning rows", err)
	}

	totalWinnings := 0
	for i := 0; i < 10000000; i++ {
		seen := make(map[int]bool)
		var currentDraw []int

		if g.SpecialBall {
			for len(currentDraw) < g.NumOfBalls-1 {
				pick := weightedWhiteList[rand.IntN(len(weightedWhiteList))]
				if !seen[pick] {
					currentDraw = append(currentDraw, pick)
					seen[pick] = true
				}
			}
			specialPick := weightSpecialList[rand.IntN(len(weightSpecialList))]
			currentDraw = append(currentDraw, specialPick)
		} else {
			for len(currentDraw) < g.NumOfBalls {
				pick := weightedWhiteList[rand.IntN(len(weightedWhiteList))]
				if !seen[pick] {
					currentDraw = append(currentDraw, pick)
					seen[pick] = true
				}
			}
		}

		// currentDraw = []int{28, 12, 52, 18, 48, 5}
		// fmt.Println(recentDraw[:len(recentDraw)-1], currentDraw[:len(currentDraw)-1])
		winner := false
		whiteBallCount := 0
		specialBallDrawn := false
		if g.SpecialBall {
			for _, num := range currentDraw[:len(currentDraw)-1] {
				if _, found := slices.BinarySearch(recentDraw[:len(recentDraw)-1], num); found {
					whiteBallCount++
				}
			}
			if currentDraw[len(currentDraw)-1] == recentDraw[len(recentDraw)-1] {
				specialBallDrawn = true
			}
		} else {
			for _, num := range currentDraw {
				if _, found := slices.BinarySearch(recentDraw, num); found {
					whiteBallCount++
				}
			}
		}

		result := fmt.Sprintf("%d:%t", whiteBallCount, specialBallDrawn)
		if _, exists := g.WaysToWin[result]; exists {
			totalWinnings += g.WaysToWin[result]
			if g.WaysToWin[result] == 0 {
				winner = true
			}
		}

		if winner {
			s += "We found a winner!!\n"

			totalCost := humanize.Comma(int64(g.CostPerPlay * float64(i)))
			count := humanize.Comma(int64(i + 1))
			winnings := humanize.Comma(int64(totalWinnings))
			profit := humanize.Comma(int64(float64(totalWinnings) - g.CostPerPlay*float64(i)))
			loss := humanize.Comma(int64(g.CostPerPlay*float64(i) - float64(totalWinnings)))

			s += fmt.Sprintf("Monty bought %s tickets for a total cost of $%s\n", count, totalCost)
			s += fmt.Sprintf("On the way to the grand prize, Monty also won $%s\n", winnings)

			if g.CostPerPlay*float64(i)-float64(totalWinnings) > 0 {
				s += fmt.Sprintf("For a total cost to win of $%s", loss)
			} else {
				s += fmt.Sprintf("For an additional profit of %s", profit)
			}

			return s
		}
	}
	s += "After 10 MILLION runs. You did not win the grand prize\n"
	totalCost := humanize.Comma(int64(10000000.00 * g.CostPerPlay))
	winnings := humanize.Comma(int64(totalWinnings))

	s += fmt.Sprintf("On the way to spending $%s.\nYou won $%s for matching some numbers", totalCost, winnings)
	return s
}

func ReviewStats(gameName string, db *sql.DB) string {
	g := game[gameName]

	whiteBallMap, specialBallMap := getData(g, db)

	makeBarChart(whiteBallMap, specialBallMap, g, "Count")

	whiteBallProbabilities, specialBallProbabilities := Probabilities(whiteBallMap, specialBallMap, g, db)
	makeBarChart(whiteBallProbabilities, specialBallProbabilities, g, "Probability")

	s := fmt.Sprintf("Finished creating graphs for %s\nThey can be found in the graphs folder", g.Name)
	return s
}
