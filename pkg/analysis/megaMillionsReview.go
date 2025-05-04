package analysis

import (
	"database/sql"
	"fmt"
	"io"
	"strings"
)

func AnalyzeMegaMillions(db *sql.DB, w io.Writer, option string) {
	whiteBalls := make(map[int]int)
	megaBall := make(map[int]int)

	for num := 1; num < 71; num++ {
		whiteBalls[num] = 0
	}
	for num := 1; num < 25; num++ {
		megaBall[num] = 0
	}
	numbersByColumn(db, "megaMillion", "N1", whiteBalls)
	numbersByColumn(db, "megaMillion", "N2", whiteBalls)
	numbersByColumn(db, "megaMillion", "N3", whiteBalls)
	numbersByColumn(db, "megaMillion", "N4", whiteBalls)
	numbersByColumn(db, "megaMillion", "N5", whiteBalls)
	numbersByColumn(db, "megaMillion", "MegaBall", megaBall)

	switch option {
	case "Counts":
		barGraph(whiteBalls, "Mega Millions White Ball Count", "White Ball Count", w)
		barGraph(megaBall, "Mega Millions Mega Ball Count", "Mega Ball Count", w)
	case "Probabilities":
		whiteBallProbability := make(map[int]float64)
		megaBallProbability := make(map[int]float64)
		findProbabilities(whiteBalls, whiteBallProbability)
		findProbabilities(megaBall, megaBallProbability)
		barGraph(whiteBallProbability, "Mega Millions White Ball Probability", "Probability %", w)
		barGraph(megaBallProbability, "Mega Millions Mega Ball Probability", "Probability %", w)
	case "Top5":
		top5White := topNNumbers(whiteBalls, 5)
		top5Mega := topNNumbers(megaBall, 5)
		var whiteList strings.Builder

		whiteList.WriteString(fmt.Sprintf("%-15s %-15s %-15s %-15s\n", "White Ball", "Count", "Mega Ball", "Count"))
		for idx, ball := range top5White {
			whiteList.WriteString(fmt.Sprintf("%-15d %-15d %-15d %-15d\n", ball[0], ball[1], top5Mega[idx][0], top5Mega[idx][1]))
		}
		fmt.Fprintln(w, whiteList.String())
	}
}
