package main

import "github.com/gdamore/tcell/v2"

var MENU_OPTIONS = []string{
	"Sprint",
	"Endless",
	"Survival",
	"Cheese",
	"Score Attack",
	"Replays",
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
	case Quit:
		ms.app.WillQuit = true
	}
}

func (ms *MenuScene) Update() {
}

func (ms *MenuScene) ConfirmAction() {
	switch ms.menuFocus {
	case 0:
		ms.app.OpenPreGameScene(
			DefaultTetrisSettings,
			LineClear,
			&LineClearSettings{
				Lines: 40,
			},
		)
	case 1:
		ms.app.OpenPreGameScene(
			DefaultTetrisSettings,
			Endless,
			&EndlessSettings{},
		)
	case 2:
		ms.app.OpenPreGameScene(
			DefaultTetrisSettings,
			Survival,
			&SurvivalSettings{
				GarbageRate: 1000,
			},
		)
	case 3:
		ms.app.OpenPreGameScene(
			DefaultTetrisSettings,
			Cheese,
			&CheeseSettings{
				Garbage: 18,
			},
		)
	case 4:
		ms.app.OpenPreGameScene(
			DefaultTetrisSettings,
			ScoreAttack,
			&ScoreAttackSettings{
				Duration: 120,
			},
		)
	case 5:
		ms.app.OpenReplayBrowserScene()
	case 6:
		break
	case 7:
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

func (ms *MenuScene) Cleanup() {
}
