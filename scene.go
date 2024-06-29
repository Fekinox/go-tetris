package main

import "github.com/gdamore/tcell/v2"

type Scene interface {
	HandleEvent(evt tcell.Event)
	HandleAction(act Action)
	Update()
	Draw(sw, sh int, rr Area, lag float64)
	Cleanup()
}

type NullScene struct {
}

func (ns *NullScene) HandleEvent(evt tcell.Event) {
}

func (ns *NullScene) HandleAction(act Action) {
}

func (ns *NullScene) Update() {
}

func (ns *NullScene) Draw(sw, sh int, rr Area, lag float64) {
}

func (ns *NullScene) Cleanup() {
}
