package main

import (
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
)

type App struct {
	CurrentScene Scene
	NextScene    Scene

	HasNextScene bool
	WillQuit     bool

	lastRenderDuration float64
	DefaultStyle       tcell.Style
}

func NewApp() *App {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	Screen = s
	if err := Screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	Screen.SetStyle(defStyle)
	Screen.EnableMouse()
	Screen.EnablePaste()
	Screen.Clear()

	app := &App{
		DefaultStyle: tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset),
	}

	app.OpenMenuScene()

	return app
}

func (a *App) Quit() {
	maybePanic := recover()
	Screen.Fini()
	if maybePanic != nil {
		panic(maybePanic)
	}
}

func (a *App) Loop() {
	lag := 0.0
	prevTime := time.Now()

	for {
		currTime := time.Now()
		elapsed := float64(currTime.Sub(prevTime).Nanoseconds()) / (1000 * 1000)
		lag += elapsed
		prevTime = currTime

		if a.NextScene != nil {
			a.CurrentScene = a.NextScene
			a.NextScene = nil
		}

		if a.WillQuit {
			return
		}

		// Event handling
		for Screen.HasPendingEvent() {
			ev := Screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventResize:
				Screen.Sync()
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
					return
				} else if ev.Key() == tcell.KeyCtrlL {
					Screen.Sync()
				} else {
					a.CurrentScene.HandleEvent(ev)
				}
			default:
				a.CurrentScene.HandleEvent(ev)
			}
		}

		dirty := false
		for lag >= UPDATE_TICK_RATE_MS {
			dirty = true
			a.CurrentScene.Update()
			lag -= UPDATE_TICK_RATE_MS
		}

		if dirty {
			a.Draw(lag)
		}
	}
}

func (a *App) Draw(lag float64) {
	Screen.Clear()

	sw, sh := Screen.Size()
	if sw < MIN_WIDTH || sh < MIN_HEIGHT {
		ShowResizeScreen(sw, sh, defStyle)
		Screen.Show()
		return
	}

	rr := Area{
		X:      (sw - MIN_WIDTH) / 2,
		Y:      (sh - MIN_HEIGHT) / 2,
		Width:  MIN_WIDTH,
		Height: MIN_HEIGHT,
	}

	BorderBox(Area{
		X:      rr.X - 1,
		Y:      rr.Y - 1,
		Width:  rr.Width + 2,
		Height: rr.Height + 2,
	}, defStyle)

	a.CurrentScene.Draw(sw, sh, rr, lag)
	Screen.Show()
}

func (a *App) OpenMenuScene() {
	menuScene := MenuScene{}
	menuScene.Init(a)

	a.NextScene = &menuScene
}

func (a *App) OpenLineClearScene() {
	gameScene := LineClearScene{}
	gameScene.Init(a, 40, 1)

	a.CurrentScene = &gameScene
}

func (a *App) OpenEndlessScene() {
	gameScene := EndlessScene{}
	gameScene.Init(a, 40, 1)

	a.CurrentScene = &gameScene
}

func (a *App) OpenSurvivalScene() {
	gameScene := SurvivalScene{}
	gameScene.Init(a, 1)

	a.CurrentScene = &gameScene
}

func (a *App) OpenCheeseScene() {
	gameScene := CheeseScene{}
	gameScene.Init(a, 1, 18)

	a.CurrentScene = &gameScene
}
