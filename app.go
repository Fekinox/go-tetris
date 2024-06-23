package main

import (
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

const TIME_SCALE float64 = 200

type App struct {
	CurrentScene Scene
	NextScene    Scene

	HasNextScene bool
	WillQuit     bool

	lastRenderDuration float64
	DefaultStyle       tcell.Style

	keyActionMap  map[tcell.Key]Action
	runeActionMap map[rune]Action

	LogFileHandle *os.File
	Logger        *log.Logger
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
		DefaultStyle: tcell.StyleDefault.Background(tcell.ColorReset).
			Foreground(tcell.ColorReset),
		keyActionMap:  make(map[tcell.Key]Action),
		runeActionMap: make(map[rune]Action),
	}

	app.keyActionMap[tcell.KeyLeft] = MoveLeft
	app.keyActionMap[tcell.KeyRight] = MoveRight
	app.keyActionMap[tcell.KeyUp] = MoveUp
	app.keyActionMap[tcell.KeyDown] = MoveDown
	app.keyActionMap[tcell.KeyEnter] = MenuConfirm

	app.runeActionMap[' '] = HardDrop
	app.runeActionMap['f'] = ToggleSuper
	app.runeActionMap['F'] = ToggleSuper

	app.runeActionMap['z'] = RotateCCW
	app.runeActionMap['Z'] = RotateCCW
	app.runeActionMap['x'] = RotateCW
	app.runeActionMap['X'] = RotateCW
	app.runeActionMap['c'] = SwapHoldPiece
	app.runeActionMap['C'] = SwapHoldPiece

	app.runeActionMap['q'] = Quit
	app.runeActionMap['Q'] = Quit
	app.runeActionMap['r'] = Reset
	app.runeActionMap['R'] = Reset
	app.runeActionMap['p'] = Pause
	app.runeActionMap['P'] = Pause

	app.OpenMenuScene()

	// Initialize logger
	app.LogFileHandle, err = os.Create("logfile")
	if err != nil {
		log.Fatalf("%+v", err)
	}

	app.Logger = log.New(app.LogFileHandle, "", log.Flags())

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
		lag += elapsed * TIME_SCALE
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
					var action Action
					var ok bool
					if ev.Key() == tcell.KeyRune {
						action, ok = a.runeActionMap[ev.Rune()]
						if !ok {
							continue
						}
					} else {
						action, ok = a.keyActionMap[ev.Key()]
						if !ok {
							continue
						}
					}
					a.CurrentScene.HandleAction(action)

					a.Logger.Printf("action: %v", action.ToString())
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
	gameScene := GameScene{}
	gameScene.Init(
		a,
		GlobalTetrisSettings{
			StartingLevel: 1,
		},
		LineClear,
		&LineClearSettings{
			Lines: 40,
		},
	)

	a.CurrentScene = &gameScene
}

func (a *App) OpenEndlessScene() {
	gameScene := GameScene{}
	gameScene.Init(
		a,
		GlobalTetrisSettings{
			StartingLevel: 1,
		},
		Endless,
		&EndlessSettings{},
	)

	a.CurrentScene = &gameScene
}

func (a *App) OpenSurvivalScene() {
	gameScene := GameScene{}
	gameScene.Init(
		a,
		GlobalTetrisSettings{
			StartingLevel: 1,
		},
		Survival,
		&SurvivalSettings{
			GarbageRate: 1000.0,
		},
	)

	a.CurrentScene = &gameScene
}

func (a *App) OpenCheeseScene() {
	gameScene := GameScene{}
	gameScene.Init(
		a,
		GlobalTetrisSettings{
			StartingLevel: 1,
		},
		Cheese,
		&CheeseSettings{
			Garbage: 18,
		},
	)

	a.CurrentScene = &gameScene
}

func (a *App) OpenReplayMenuScene() {
	menuScene := ReplayBrowserScene{}
	menuScene.Init(a)
	a.CurrentScene = &menuScene
}

func (a *App) OpenReplayScene(data ReplayData) {
	replayScene := ReplayViewerScene{}
	replayScene.Init(
		a,
		data,
	)

	a.CurrentScene = &replayScene
}
