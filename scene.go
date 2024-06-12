package main

import "github.com/gdamore/tcell/v2"

type Scene interface {
	Init(app *App)
	HandleEvent(evt tcell.Event)
	HandleAction(act Action)
	Update()
	Draw(lag float64)
}
