package main

import "github.com/gdamore/tcell/v2"

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
	lcs.es.Update()
}

func (lcs *LineClearScene) Draw(sw, sh int, rr Area, lag float64) {
	lcs.es.Draw(sw, sh, rr, lag)
}
