package main

import "github.com/gdamore/tcell/v2"

type Scene interface {
	HandleEvent(evt tcell.Event)
	HandleAction(act Action)
	Update()
	Draw(sw, sh int, rr Area, lag float64)
	Cleanup()
}
