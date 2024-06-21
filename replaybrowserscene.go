package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
)

type ReplayBrowserScene struct {
	app *App

	replayFileNames []string
	loaded          bool
	menuFocus       int
}

func (ms *ReplayBrowserScene) Init(app *App) {
	ms.app = app

	go func() {
		replayDir, err := os.Open("replays")
		if err != nil {
			return

		}
		entries, err := replayDir.ReadDir(0)
		if err != nil {
			return
		}
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}

		ms.replayFileNames = names
		ms.loaded = true
	}()
}

func (ms *ReplayBrowserScene) HandleEvent(evt tcell.Event) {
}

func (ms *ReplayBrowserScene) HandleAction(act Action) {
	switch act {
	case Quit:
		ms.app.OpenMenuScene()
	case MoveUp:
		if ms.loaded {
			ms.menuFocus = max(0, ms.menuFocus-1)
		}
	case MoveDown:
		if ms.loaded {
			ms.menuFocus = min(len(ms.replayFileNames), ms.menuFocus+1)
		}
	case MenuConfirm:
		if ms.loaded {
			ms.ConfirmAction()
		}
	}
}

func (ms *ReplayBrowserScene) Update() {
}

func (ms *ReplayBrowserScene) ConfirmAction() {
	name := ms.replayFileNames[ms.menuFocus]
	file, err := os.Open(fmt.Sprintf("replays/%s", name))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	replayData := ReplayData{}
	err = replayData.DecodeCompressed(file)

	if err != nil {
		panic(err)
	}

	ms.app.OpenReplayScene(replayData)
}

func (ms *ReplayBrowserScene) Draw(sw, sh int, rr Area, lag float64) {
	SetString(
		rr.X,
		rr.Y,
		"Replays",
		defStyle)

	if ms.loaded {
		for i, opt := range ms.replayFileNames {
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
}
