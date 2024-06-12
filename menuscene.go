package main

import "github.com/gdamore/tcell/v2"

var MENU_OPTIONS = []string{
	"Play",
	"Credits",
	"Quit",
}

type MenuScene struct {
	app *App

	menuFocus int
}

func (ms *MenuScene) Init(app *App) {
	ms.app = app
}

func (ms *MenuScene) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyUp || ev.Key() == tcell.KeyLeft {
			ms.menuFocus = max(0, ms.menuFocus-1)
		}
		if ev.Key() == tcell.KeyDown || ev.Key() == tcell.KeyRight {
			ms.menuFocus = min(len(MENU_OPTIONS), ms.menuFocus+1)
		}
		if ev.Key() == tcell.KeyEnter || IsRune(ev, ' ') {
			ms.ConfirmAction()
		}
	}
}

func (ms *MenuScene) HandleAction(act Action) {
}

func (ms *MenuScene) Update() {
}

func (ms *MenuScene) ConfirmAction() {
	switch ms.menuFocus {
	case 0:
		ms.app.OpenGameScene()
	case 1:
		break;
	case 2:
		ms.app.WillQuit = true
	}
}

func (ms *MenuScene) Draw(sw, sh int, rr Area, lag float64) {
	SetString(
		rr.X,
		rr.Y,
		"Tetris",
		defStyle)

	for i, opt := range MENU_OPTIONS {
		style := defStyle
		if i == ms.menuFocus {
			style = style.Reverse(true)
			Screen.SetContent(
				rr.X,
				rr.Y + 2 + 2 * i,
				'*',
				nil, defStyle)
		}

		SetString(
			rr.X + 2,
			rr.Y + 2 + 2 * i,
			opt,
			style)
	}
}
