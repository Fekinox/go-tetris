package main

import (
	"github.com/gdamore/tcell/v2"
)

const MAX_CHEESE_GARBAGE_LINES = 10

type CheeseScene struct {
	app           *App
	es            *TetrisField
	startingLevel int64

	startingGarbage int
	currentGarbage  int
}

func (chs *CheeseScene) Init(app *App, level int64, startingGarbage int) {
	chs.app = app
	chs.es = NewTetrisField(level)
	chs.startingGarbage = startingGarbage

	chs.currentGarbage = chs.startingGarbage

	for i := 0; i < chs.currentGarbage && i < MAX_CHEESE_GARBAGE_LINES; i++ {
		chs.es.AddGarbage(1)
	}

	chs.es.AddLineClearHandler(func(garbage, nonGarbage int) {
		chs.OnLineClear(garbage)
	})
}

func (chs *CheeseScene) HandleEvent(ev tcell.Event) {
}

func (chs *CheeseScene) HandleAction(act Action) {
	switch act {
	case Quit:
		chs.app.OpenMenuScene()
	case Reset:
		chs.es.HandleReset()
		chs.currentGarbage = chs.startingGarbage

		for i := 0; i < chs.currentGarbage && i < MAX_CHEESE_GARBAGE_LINES; i++ {
			chs.es.QueueGarbage(1)
		}
	default:
		chs.es.HandleAction(act)
	}
}

func (chs *CheeseScene) OnLineClear(garbage int) {
	for i := 0; i < garbage; i++ {
		chs.currentGarbage -= 1
		if chs.currentGarbage == 0 {
			chs.es.gameOver = true
		}
		if chs.currentGarbage >= MAX_CHEESE_GARBAGE_LINES {
			chs.es.QueueGarbage(1)
		}
	}
}

func (chs *CheeseScene) Update() {
	if chs.es.gameOver {
		return
	}

	chs.es.Update()
}

func (chs *CheeseScene) Draw(sw, sh int, rr Area, lag float64) {
	playingField := rr.Inset(BOARD_WIDTH, BOARD_HEIGHT+4)
	anchorX := playingField.X - 2
	anchorY := playingField.Bottom() - 2

	chs.es.Draw(sw, sh, playingField, lag)
	chs.es.DrawStats(rr, anchorX, anchorY)
}
