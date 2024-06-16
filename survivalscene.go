package main

import (
	"github.com/gdamore/tcell/v2"
)

const SURVIVAL_GARBAGE_RATE = 60

type SurvivalScene struct {
	app           *App
	es            *TetrisField
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
			svs.survivalCounter = SURVIVAL_GARBAGE_RATE
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
	playingField := rr.Inset(BOARD_WIDTH, BOARD_HEIGHT+4)
	anchorX := playingField.X - 2
	anchorY := playingField.Bottom() - 2

	svs.es.Draw(sw, sh, playingField, lag)
	svs.es.DrawStats(rr, anchorX, anchorY)
}