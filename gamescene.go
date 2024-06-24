package main

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

const COUNTDOWN_DURATION_SECS = 4.0

var COUNTDOWN_TIMER_LEVELS = []rune(
	" .-*-. ",
)

type GameScene struct {
	app *App
	es  *TetrisField

	seed              int64
	globalSettings    GlobalTetrisSettings
	objectiveID       ObjectiveID
	objectiveSettings ObjectiveSettings
	objective         Objective

	countdownTimer float64
	gameStarted    bool

	actions []ReplayAction

	stats []Stat
}

func (gs *GameScene) Init(
	app *App,
	globalSettings GlobalTetrisSettings,
	objectiveID ObjectiveID,
	objectiveSettings ObjectiveSettings,
) {
	gs.app = app
	gs.seed = time.Now().UnixNano()
	gs.es = NewTetrisField(gs.seed, globalSettings)

	gs.globalSettings = globalSettings
	gs.objectiveID = objectiveID
	gs.objectiveSettings = objectiveSettings
	gs.objective = gs.objectiveSettings.Init(gs.es)

	gs.countdownTimer = COUNTDOWN_DURATION_SECS
	gs.gameStarted = false

	gs.actions = make([]ReplayAction, 0)

	gs.es.AddGameOverHandler(func(failed bool, reason string) {
		gs.OnGameOver(failed, reason)
	})

	gs.stats = []Stat {
		CreateElapsedTimeStat(gs.es),
		CreatePiecesStat(gs.es),
		CreateLinesStat(gs.es),
	}
}

func (gs *GameScene) HandleEvent(ev tcell.Event) {
}

func (gs *GameScene) HandleAction(act Action) {
	switch act {
	case Quit:
		gs.app.OpenMenuScene()
	case Reset:
		gs.seed = time.Now().UnixNano()
		gs.es.HandleReset(gs.seed)
		gs.objective = gs.objectiveSettings.Init(gs.es)

		gs.countdownTimer = COUNTDOWN_DURATION_SECS
		gs.gameStarted = false

		gs.actions = make([]ReplayAction, 0)

		gs.es.AddGameOverHandler(func(failed bool, reason string) {
			gs.OnGameOver(failed, reason)
		})
	default:
		if gs.gameStarted {
			gs.actions = append(gs.actions, ReplayAction{
				Action: act,
				Frame:  gs.es.frameCount,
			})
			gs.objective.HandleAction(act, gs.es)
		}
	}
}

func (gs *GameScene) Update() {
	if !gs.gameStarted {
		gs.countdownTimer -= UPDATE_TICK_RATE_MS / 1000.0
		if gs.countdownTimer < 0 {
			gs.gameStarted = true
			gs.es.gameStarted = true
			gs.es.GetRandomPiece()
		}

		return
	}

	gs.objective.Update(gs.es)
}

func (gs *GameScene) OnGameOver(failed bool, reason string) {
	replayData := ReplayData{
		Seed:              gs.seed,
		TetrisSettings:    gs.globalSettings,
		ObjectiveID:       gs.objectiveID,
		ObjectiveSettings: gs.objectiveSettings,
		Actions:           gs.actions,
	}

	gs.app.Logger.Printf("Seed: %v\n", gs.seed)
	gs.app.Logger.Printf("Settings: %v\n", gs.globalSettings)
	gs.app.Logger.Printf("ObjectiveID: %v\n", gs.objectiveID)
	gs.app.Logger.Printf("ObjectiveSettings: %v\n", gs.objectiveSettings)
	gs.app.Logger.Printf("Number of actions: %v\n", len(gs.actions))

	err := os.Mkdir("replays", 0755)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		panic(err)
	}

	file, err := os.Create(fmt.Sprintf("replays/rp-%v", time.Now()))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = StdEncoder(&replayData, file)
	if err != nil {
		panic(err)
	}
}

func (gs *GameScene) Draw(sw, sh int, rr Area, lag float64) {
	playingField := rr.Inset(BOARD_WIDTH, BOARD_HEIGHT+4)
	anchorX := playingField.X - 2
	anchorY := playingField.Bottom() - 2

	gs.es.Draw(sw, sh, playingField, lag)

	yOffset := 0
	for _, stat := range gs.stats {
		strings := stat.Compute()
		SetStringArray(
			anchorX,
			anchorY+yOffset - len(strings),
			defStyle,
			true,
			strings...,
		)
		yOffset -= len(strings) + 1
	}

	if !gs.gameStarted {
		textAnchorX := playingField.X + BOARD_WIDTH/2
		textAnchorY := playingField.Y + 4
		var theText string
		if gs.countdownTimer > 3.0 {
			theText = "3..."
		} else if gs.countdownTimer > 2.0 {
			theText = "2..."
		} else if gs.countdownTimer > 1.0 {
			theText = "1..."
		} else {
			theText = "GO!!"
		}

		SetCenteredString(textAnchorX, textAnchorY, theText, defStyle)
		gs.DrawProgressBar(
			textAnchorX, textAnchorY+1,
			gs.countdownTimer-math.Floor(gs.countdownTimer),
		)
	}
}

func (gs *GameScene) DrawProgressBar(anchorX, anchorY int, value float64) {
	for i := 0; i < BOARD_WIDTH; i++ {
		intensity := value*10 - float64(i)
		intIntensity := max(
			0,
			min(
				len(COUNTDOWN_TIMER_LEVELS)-1,
				int(0.25*intensity*float64(len(PARTICLE_LEVELS))),
			),
		)
		Screen.SetContent(
			anchorX+i-BOARD_WIDTH/2,
			anchorY,
			COUNTDOWN_TIMER_LEVELS[intIntensity],
			nil, defStyle,
		)
	}
}
