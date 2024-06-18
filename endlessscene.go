package main

import (
	"github.com/gdamore/tcell/v2"
)

type EndlessScene struct {
	app           *App
	es            *TetrisField
	startingLevel int64
}

func (els *EndlessScene) Init(app *App, lineLimit int64, level int64) {
	els.app = app
	els.es = NewTetrisField(level)
}

func (els *EndlessScene) HandleEvent(ev tcell.Event) {
}

func (els *EndlessScene) HandleAction(act Action) {
	if act == Quit {
		els.app.OpenMenuScene()
		return
	}
	els.es.HandleAction(act)
}

func (els *EndlessScene) Update() {
	if !els.es.gameOver {
		els.es.Update()
	}
}

func (els *EndlessScene) Draw(sw, sh int, rr Area, lag float64) {
	// Center the playing field
	playingField := rr.Inset(BOARD_WIDTH, BOARD_HEIGHT+4)
	anchorX := playingField.X - 2
	anchorY := playingField.Bottom() - 2

	els.es.Draw(sw, sh, playingField, lag)
	els.es.DrawStats(rr, anchorX, anchorY)
}
