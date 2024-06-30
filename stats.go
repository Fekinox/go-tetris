package main

import (
	"fmt"
	"math"
)

type Stat struct {
	Name    string
	Compute func() []string
}

func DrawStats(stats []Stat, x, y int) {
	yOffset := 0
	for _, stat := range stats {
		strings := stat.Compute()
		SetStringArray(
			x,
			y+yOffset-len(strings),
			defStyle,
			true,
			strings...,
		)
		yOffset -= len(strings) + 1
	}
}

func CreateLinesStat(es *TetrisField) Stat {
	return Stat{
		Compute: func() []string {
			rawTime := float64(es.frameCount) * UPDATE_TICK_RATE_MS

			lpm := float64(es.lines) / (rawTime / (1000))

			if math.IsNaN(lpm) || math.IsInf(lpm, 0) {
				lpm = 0
			}

			linesPerMinute := fmt.Sprintf("%.2f l/m", lpm)

			return []string{
				"LINES",
				fmt.Sprintf("%d", es.lines),
				linesPerMinute,
			}
		},
	}
}

func CreateLinesRemainingStat(es *TetrisField, limit int64) Stat {
	return Stat{
		Compute: func() []string {
			rawTime := float64(es.frameCount) * UPDATE_TICK_RATE_MS

			lpm := float64(es.lines) / (rawTime / (1000))

			if math.IsNaN(lpm) || math.IsInf(lpm, 0) {
				lpm = 0
			}

			linesPerMinute := fmt.Sprintf("%.2f l/m", lpm)

			return []string{
				"LINES",
				fmt.Sprintf("%d/%d", es.lines, limit),
				linesPerMinute,
			}
		},
	}
}

func CreatePiecesStat(es *TetrisField) Stat {
	return Stat{
		Compute: func() []string {
			rawTime := float64(es.frameCount) * UPDATE_TICK_RATE_MS

			pps := float64(es.pieceCount) / (rawTime / (1000))

			if math.IsNaN(pps) || math.IsInf(pps, 0) {
				pps = 0
			}

			piecesPerSecond := fmt.Sprintf("%.2f p/s", pps)

			return []string{
				"PIECES",
				fmt.Sprintf("%d", es.pieceCount),
				piecesPerSecond,
			}
		},
	}
}

func CreateElapsedTimeStat(es *TetrisField) Stat {
	return Stat{
		Compute: func() []string {
			rawTime := float64(es.frameCount) * UPDATE_TICK_RATE_MS
			timeMinutes := math.Trunc(rawTime / (60 * 1000))
			timeSeconds := math.Trunc((rawTime - timeMinutes*60*1000) / 1000)
			timeMillis := math.Trunc((rawTime - timeMinutes*60*1000 -
				timeSeconds*1000))

			return []string{
				"TIME",
				fmt.Sprintf("%0d:%02d.%03d",
					int(timeMinutes),
					int(timeSeconds),
					int(timeMillis),
				),
			}
		},
	}
}

func CreateCountdownStat(es *TetrisField, duration int64) Stat {
	return Stat{
		Compute: func() []string {
			rawTime := float64(es.frameCount) * UPDATE_TICK_RATE_MS
			rawTime = float64(duration)*1000 - rawTime
			timeMinutes := math.Trunc(rawTime / (60 * 1000))
			timeSeconds := math.Trunc((rawTime - timeMinutes*60*1000) / 1000)
			timeMillis := math.Trunc((rawTime - timeMinutes*60*1000 -
				timeSeconds*1000))

			return []string{
				"TIME",
				fmt.Sprintf("%0d:%02d.%03d",
					int(timeMinutes),
					int(timeSeconds),
					int(timeMillis),
				),
			}
		},
	}
}

func CreateGarbageStat(co *CheeseObjective) {
}
