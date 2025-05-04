package analysis

import (
	"database/sql"
	"fmt"
	"io"
	"strings"
)

func AnalyzePowerball(db *sql.DB, w io.Writer, option string) {
	whiteBalls := make(map[int]int)
	powerBall := make(map[int]int)

	for num := 1; num < 70; num++ {
		whiteBalls[num] = 0
		if num < 27 {
			powerBall[num] = 0
		}
	}

	numbersByColumn(db, "powerBall", "WhiteBall1", whiteBalls)
	numbersByColumn(db, "powerBall", "WhiteBall2", whiteBalls)
	numbersByColumn(db, "powerBall", "WhiteBall3", whiteBalls)
	numbersByColumn(db, "powerBall", "WhiteBall4", whiteBalls)
	numbersByColumn(db, "powerBall", "WhiteBall5", whiteBalls)
	numbersByColumn(db, "powerBall", "PowerBall", powerBall)

	switch option {
	case "Counts":
		barGraph(whiteBalls, "Powerball White Ball Count", "White Ball Count", w)
		barGraph(powerBall, "Powerball Mega Ball Count", "Mega Ball Count", w)
	case "Probabilities":
		whiteBallProbability := make(map[int]float64)
		powerBallProbability := make(map[int]float64)
		findProbabilities(whiteBalls, whiteBallProbability)
		findProbabilities(powerBall, powerBallProbability)
		barGraph(whiteBallProbability, "Powerball White Ball Probability", "Probability %", w)
		barGraph(powerBallProbability, "Powerball Mega Ball Probability", "Probability %", w)
	case "Top5":
		top5White := topNNumbers(whiteBalls, 5)
		top5Power := topNNumbers(powerBall, 5)
		var whiteList strings.Builder

		whiteList.WriteString(fmt.Sprintf("%-15s %-15s %-15s %-15s\n", "White Ball", "Count", "Powerball", "Count"))
		for idx, ball := range top5White {
			whiteList.WriteString(fmt.Sprintf("%-15d %-15d %-15d %-15d\n", ball[0], ball[1], top5Power[idx][0], top5Power[idx][1]))
		}
		fmt.Fprintln(w, whiteList.String())
	}

}
