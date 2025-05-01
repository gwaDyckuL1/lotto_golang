package analysis

import (
	"database/sql"
	"fmt"
)

func AnalyzeMegaMillions(db *sql.DB) {
	whiteBalls := make(map[int]int)
	megaBall := make(map[int]int)
	whiteBallProbability := make(map[int]float64)
	megaBallProbability := make(map[int]float64)

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

	findProbabilities(whiteBalls, whiteBallProbability)
	findProbabilities(megaBall, megaBallProbability)

	barGraph(whiteBalls, "Mega Millions White Ball Count", "White Ball Count")
	barGraph(megaBall, "Mega Millions Mega Ball Count", "Mega Ball Count")
	barGraph(whiteBallProbability, "Mega Millions White Ball Probability", "Probability %")
	barGraph(megaBallProbability, "Mega Millions Mega Ball Probability", "Probability %")

	fmt.Println(topNNumbers(whiteBalls, 5))
}
