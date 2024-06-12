package main

import "github.com/gdamore/tcell/v2"

type GameScene struct {
	app *App
	es *EngineState
}

func (gs *GameScene) Init(app *App) {
	gs.app = app
	gs.es = NewEngineState()
}

func (gs *GameScene) HandleEvent(evt tcell.Event) {
	gs.es.HandleInput(evt)
}

func (gs *GameScene) HandleAction(act Action) {
	//
}

func (gs *GameScene) Update() {
	gs.es.Update()
}

func (gs *GameScene) Draw(lag float64) {
	gs.es.Draw(lag)
}
