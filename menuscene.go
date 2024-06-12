package main

import "github.com/gdamore/tcell/v2"

type MenuScene struct {
	app *App
}

func (ms*MenuScene) Init(app *App) {
}

func (gs *MenuScene) HandleEvent(evt tcell.Event) {
}

func (gs *MenuScene) HandleAction(act Action) {
}

func (gs *MenuScene) Update() {
}

func (gs *MenuScene) Draw(sw, sh int, lag float64) {
}
