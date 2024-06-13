package main

import (
	"fmt"
	"math"

	"github.com/gdamore/tcell/v2"
)

type EndlessScene struct {
	app *App
	es *TetrisField
	startingLevel int64
}

func (els *EndlessScene) Init(app *App, lineLimit int64, level int64) {
	els.app = app
	els.es = NewTetrisField(level)
}

func (els *EndlessScene) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if IsRune(ev, 'q') || IsRune(ev, 'Q') {
			els.app.OpenMenuScene()
		}
		els.es.HandleInput(ev)
	}
}

func (els *EndlessScene) HandleAction(act Action) {
	//
}

func (els *EndlessScene) Update() {
	if !els.es.gameOver {
		els.es.Update()
	}
}

func (els *EndlessScene) Draw(sw, sh int, rr Area, lag float64) {
	// Center the playing field
	playingField := rr.Inset(BOARD_WIDTH, BOARD_HEIGHT+4)
	els.es.Draw(sw, sh, playingField, lag)

	// Draw the current time
	lowerRightHudAnchorX := playingField.X - 2
	lowerRightHudAnchorY := playingField.Bottom() - 2

	rawTime := float64(els.es.frameCount) * UPDATE_TICK_RATE_MS
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
		els.es.pieceCount,
	)

	piecesPerSecondString := fmt.Sprintf(
		"%.2f p/s",
		float64(els.es.pieceCount)/(rawTime/1000),
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
		els.es.lines)

	linesPerSecondString := fmt.Sprintf(
		"%.2f l/s",
		float64(els.es.lines)/(rawTime/1000))

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
		fmt.Sprintf("%d", els.es.score),
		defStyle)
	SetCenteredString(
		centerBottomAnchorX,
		centerBottomAnchorY+1,
		fmt.Sprintf("%d", els.es.level),
		defStyle)
}
