package main

import (
	"fmt"
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type LineClearScene struct {
	app *App
	es *TetrisField
	lineLimit int64
	startingLevel int64
}

func (lcs *LineClearScene) Init(app *App, lineLimit int64, level int64) {
	lcs.app = app
	lcs.es = NewTetrisField(level)
	lcs.lineLimit = lineLimit
}

func (lcs *LineClearScene) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if IsRune(ev, 'q') || IsRune(ev, 'Q') {
			lcs.app.OpenMenuScene()
		}
		lcs.es.HandleInput(ev)
		lcs.AfterEvent()
	}
}

func (lcs *LineClearScene) AfterEvent() {
	if lcs.es.lines >= lcs.lineLimit {
		lcs.es.gameOver = true
	}
}

func (lcs *LineClearScene) HandleAction(act Action) {
	//
}

func (lcs *LineClearScene) Update() {
	if !lcs.es.gameOver {
		lcs.es.Update()
	}
}

func (lcs *LineClearScene) Draw(sw, sh int, rr Area, lag float64) {
	// Center the playing field
	playingField := rr.Inset(BOARD_WIDTH, BOARD_HEIGHT+4)
	lcs.es.Draw(sw, sh, playingField, lag)

	// Draw the current time
	lowerRightHudAnchorX := playingField.X - 2
	lowerRightHudAnchorY := playingField.Bottom() - 2

	rawTime := float64(lcs.es.frameCount) * UPDATE_TICK_RATE_MS
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
	timeStringLength := runewidth.StringWidth(timeString)

	SetString(
		lowerRightHudAnchorX - 4,
		lowerRightHudAnchorY - 1,
		"TIME",
		defStyle)
	SetString(
		lowerRightHudAnchorX - timeStringLength,
		lowerRightHudAnchorY,
		timeString,
		defStyle)

	pieceCountString := fmt.Sprintf(
		"%d",
		lcs.es.pieceCount,
	)
	pieceCountStringLength := runewidth.StringWidth(pieceCountString)

	piecesPerSecondString := fmt.Sprintf(
		"%.2f p/s",
		float64(lcs.es.pieceCount)/(rawTime/1000),
	)

	piecesPerSecondStringLength := runewidth.StringWidth(piecesPerSecondString)

	SetString(
		lowerRightHudAnchorX - 6,
		lowerRightHudAnchorY - 5,
		"PIECES",
		defStyle)


	SetString(
		lowerRightHudAnchorX - pieceCountStringLength,
		lowerRightHudAnchorY - 4,
		pieceCountString,
		defStyle)

	SetString(
		lowerRightHudAnchorX - piecesPerSecondStringLength,
		lowerRightHudAnchorY - 3,
		piecesPerSecondString,
		defStyle)
}
