package analysis

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"sort"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func numbersByColumn(db *sql.DB, table string, columnName string, theMap map[int]int) {
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

func topNNumbers(numMap map[int]int, n int) [][]int {
	topNums := [][]int{}

	for key, val := range numMap {
		topNums = append(topNums, []int{key, val})
	}

	sort.Slice(topNums, func(i, j int) bool {
		return topNums[i][1] > topNums[j][1]
	})
	return topNums[:n]
}

func findProbabilities(numMap map[int]int, numProb map[int]float64) {
	totalDraws := 0

	for _, val := range numMap {
		totalDraws += val
	}

	for key, val := range numMap {
		numProb[key] = (float64(val) / float64(totalDraws)) * 100.0
	}
}

func monteCarlo(white map[int]int, yellow map[int]int) ([][]int, [][]int) {
	whiteProbability := make(map[int]float64)
	yellowProbability := make(map[int]float64)

	findProbabilities(white, whiteProbability)
	findProbabilities(yellow, yellowProbability)

	weightWhite := []int{}
	weightYellow := []int{}
	scaleFactor := 1000.00

	for num, prob := range whiteProbability {
		count := int(prob * scaleFactor)
		for range count {
			weightWhite = append(weightWhite, num)
		}
	}
	for num, prob := range yellowProbability {
		count := int(prob * scaleFactor)
		for range count {
			weightYellow = append(weightYellow, num)
		}
	}

	testWhiteball := make(map[int]int)
	testYellowball := make(map[int]int)

	for i := 1; i < len(white); i++ {
		testWhiteball[i] = 0
		if i < len(yellow) {
			testYellowball[i] = 0
		}
	}

	for range 10000 {
		rand.Shuffle(len(weightWhite), func(i, j int) {
			weightWhite[i], weightWhite[j] = weightWhite[j], weightWhite[i]
		})
		rand.Shuffle(len(weightYellow), func(i, j int) {
			weightYellow[i], weightYellow[j] = weightYellow[j], weightYellow[i]
		})

		testYellowball[weightYellow[0]]++

		test := []int{}
		seen := map[int]bool{}

		for _, val := range weightWhite {
			if !seen[val] {
				test = append(test, val)
				seen[val] = true
			}
			if len(test) == 5 {
				break
			}
		}
		for _, val := range test {
			testWhiteball[val]++
		}
	}
	whiteNumberstoPick := topNNumbers(testWhiteball, 5)
	yellowNumbertoPick := topNNumbers(testYellowball, 1)

	return whiteNumberstoPick, yellowNumbertoPick
}

func barGraph[T int | float64](data map[int]T, tableName string, yTitle string, w io.Writer) {

	p := plot.New()
	p.Title.Text = tableName
	p.X.Label.Text = "Number"
	p.Y.Label.Text = yTitle

	xAxis := make([]string, len(data))
	yAxis := make(plotter.Values, len(data))

	xAxisNums := []int{}
	for num := range data {
		xAxisNums = append(xAxisNums, num)
	}
	sort.Ints(xAxisNums)

	for i, v := range xAxisNums {
		xAxis[i] = fmt.Sprintf("%d", v)
		yAxis[i] = float64(data[v])
	}

	bar, err := plotter.NewBarChart(yAxis, vg.Points(10))
	if err != nil {
		log.Fatalf("Failed to create bar chart: %v", err)
	}
	bar.LineStyle.Width = vg.Length(0)
	bar.Color = plotutil.Color(2)

	p.Add(bar)
	p.NominalX(xAxis...)

	tableName = "graphs/" + tableName + ".png"
	width := vg.Length(len(data)) * vg.Points(15)
	height := vg.Inch * 6

	if err := p.Save(width, height, tableName); err != nil {
		log.Fatalf("Failed to save plot: %v", err)
	}
	fmt.Fprintln(w, "Chart saved as: ", tableName)
}
