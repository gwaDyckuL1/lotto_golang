package analysis

import (
	"database/sql"
	"io"
)

func AnalyzeWALotto(db *sql.DB, w io.Writer, option string) {
	ballCounts := make(map[int]int)

	for num := 1; num < 50; num++ {
		ballCounts[num] = 0
	}

	numbersByColumn(db, "waLotto", "Ball1", ballCounts)
	numbersByColumn(db, "waLotto", "Ball2", ballCounts)
	numbersByColumn(db, "waLotto", "Ball3", ballCounts)
	numbersByColumn(db, "waLotto", "Ball4", ballCounts)
	numbersByColumn(db, "waLotto", "Ball5", ballCounts)
	numbersByColumn(db, "waLotto", "Ball6", ballCounts)

	switch option {
	case "Counts":
	case "Probabilities":
	case "Top5":
	case "Monte Carlo":
	}
}
