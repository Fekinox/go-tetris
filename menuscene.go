package main

import "github.com/gdamore/tcell/v2"

var MENU_OPTIONS = []string{
	"40 Line Clear",
	"Endless",
	"Survival",
	"Cheese",
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
}

func (ms *MenuScene) HandleAction(act Action) {
	switch act {
		case MoveUp:
			ms.menuFocus = max(0, ms.menuFocus-1)
		case MoveDown:
			ms.menuFocus = min(len(MENU_OPTIONS), ms.menuFocus+1)
		case MenuConfirm:
			ms.ConfirmAction()
	}
}

func (ms *MenuScene) Update() {
}

func (ms *MenuScene) ConfirmAction() {
	switch ms.menuFocus {
	case 0:
		ms.app.OpenLineClearScene()
	case 1:
		ms.app.OpenEndlessScene()
	case 2:
		ms.app.OpenSurvivalScene()
	case 3:
		ms.app.OpenCheeseScene()
	case 4:
		break
	case 5:
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
				rr.Y+2+2*i,
				'*',
				nil, defStyle)
		}

		SetString(
			rr.X+2,
			rr.Y+2+2*i,
			opt,
			style)
	}
}
