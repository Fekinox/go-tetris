package main

import (
	"fmt"
	"math"

	"github.com/gdamore/tcell/v2"
)

const SURVIVAL_GARBAGE_RATE = 60

type SurvivalScene struct {
	app *App
	es *TetrisField
	startingLevel int64

	survivalCounter int
}

func (svs *SurvivalScene) Init(app *App, level int64) {
	svs.app = app
	svs.es = NewTetrisField(level)
	svs.survivalCounter = SURVIVAL_GARBAGE_RATE
}

func (svs *SurvivalScene) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if IsRune(ev, 'q') || IsRune(ev, 'Q') {
			svs.app.OpenMenuScene()
		} else if IsRune(ev, 'r') || IsRune(ev, 'R') {
			svs.es.HandleReset()
		} else {
			svs.es.HandleInput(ev)
		}
	}
}

func (svs *SurvivalScene) HandleAction(act Action) {
	//
}

func (svs *SurvivalScene) Update() {
	if svs.es.gameOver {
		return
	}

	svs.es.Update()

	svs.survivalCounter--
	if svs.survivalCounter < 0 {
		svs.survivalCounter = SURVIVAL_GARBAGE_RATE
		svs.es.AddGarbage(1)
	}
}

func (svs *SurvivalScene) Draw(sw, sh int, rr Area, lag float64) {
	// Center the playing field
	playingField := rr.Inset(BOARD_WIDTH, BOARD_HEIGHT+4)
	svs.es.Draw(sw, sh, playingField, lag)

	// Draw the current time
	lowerRightHudAnchorX := playingField.X - 2
	lowerRightHudAnchorY := playingField.Bottom() - 2

	rawTime := float64(svs.es.frameCount) * UPDATE_TICK_RATE_MS
	timeMinutes := math.Trunc(rawTime/(60*1000))
	timeSeconds := math.Trunc((rawTime - timeMinutes*60*1000)/1000)
	timeMillis := math.Trunc((rawTime - timeMinutes*60*1000 -
	timeSeconds*1000))

	timeString := fmt.Sprintf(
		"%0d:%02d.%03d",
		int64(timeMinutes),
		int64(timeSeconds),
		int64(timeMillis),
	)

	SetStringArray(
		lowerRightHudAnchorX,
		lowerRightHudAnchorY - 1,
		defStyle,
		true,
		"TIME",
		timeString)

	pieceCountString := fmt.Sprintf(
		"%d",
		svs.es.pieceCount,
	)

	piecesPerSecondString := fmt.Sprintf(
		"%.2f p/s",
		float64(svs.es.pieceCount)/(rawTime/1000),
	)

	SetStringArray(
		lowerRightHudAnchorX,
		lowerRightHudAnchorY - 5,
		defStyle,
		true,
		"PIECES",
		pieceCountString,
		piecesPerSecondString)

	// Draw lines & lines per second
	linesString := fmt.Sprintf(
		"%d",
		svs.es.lines)

	linesPerSecondString := fmt.Sprintf(
		"%.2f l/s",
		float64(svs.es.lines)/(rawTime/1000))

	SetStringArray(
		lowerRightHudAnchorX,
		lowerRightHudAnchorY - 9,
		defStyle,
		true,
		"LINES",
		linesString,
		linesPerSecondString)

	// Draw score and level
	
	centerBottomAnchorX := playingField.Left() + playingField.Width/2
	centerBottomAnchorY := playingField.Bottom() - 1
	SetCenteredString(
		centerBottomAnchorX,
		centerBottomAnchorY,
		fmt.Sprintf("%d", svs.es.score),
		defStyle)
	SetCenteredString(
		centerBottomAnchorX,
		centerBottomAnchorY+1,
		fmt.Sprintf("%d", svs.es.level),
		defStyle)
}
