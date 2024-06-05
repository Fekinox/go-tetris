package main

import (
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
)

type App struct {
	EngineState        *EngineState
	lastRenderDuration float64
	DefaultStyle       tcell.Style
}

func NewApp() App {
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

	return App {
		EngineState: InitEngineState(),
		DefaultStyle: tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset),
	}
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
	lastRender := time.Now()

	for {
		currTime := time.Now()
		elapsed := float64(currTime.Sub(prevTime).Nanoseconds()) / (1000 * 1000)
		lag += elapsed
		prevTime = currTime

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
					a.EngineState.HandleInput(ev)
				}
			default:
				a.EngineState.HandleInput(ev)
			}
		}

		dirty := false
		for lag >= UPDATE_TICK_RATE_MS {
			dirty = true
			a.EngineState.Update()
			lag -= UPDATE_TICK_RATE_MS
		}

		if dirty {
			a.EngineState.Draw(lag)
			currRender := time.Now()
			a.EngineState.LastRenderDuration = float64(currRender.Sub(lastRender).Nanoseconds()) / (1000 * 1000)
			lastRender = currRender
		}
	}
}
