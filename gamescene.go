package main

import "github.com/gdamore/tcell/v2"

type GameScene struct {
	app *App
	es *TetrisField
}

func (gs *GameScene) Init(app *App) {
	gs.app = app
	gs.es = NewTetrisField(1)
}

func (gs *GameScene) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if IsRune(ev, 'q') || IsRune(ev, 'Q') {
			gs.app.OpenMenuScene()
		}
		gs.es.HandleInput(ev)
	}
}

func (gs *GameScene) HandleAction(act Action) {
	//
}

func (gs *GameScene) Update() {
	gs.es.Update()
}

func (gs *GameScene) Draw(sw, sh int, rr Area, lag float64) {
	gs.es.Draw(sw, sh, rr, lag)
}
