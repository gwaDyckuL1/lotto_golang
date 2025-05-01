package analysis

import (
	"database/sql"
	"fmt"
	"log"
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

func barGraph[T int | float64](data map[int]T, tableName string, yTitle string) {

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
	fmt.Println("Chart saved as: ", tableName)
}
