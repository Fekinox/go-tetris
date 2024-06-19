package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

const COUNTDOWN_DURATION_SECS = 4.0

type GameScene struct {
	app *App
	es *TetrisField

	seed int64
	globalSettings GlobalTetrisSettings
	objectiveSettings ObjectiveSettings
	objective Objective

	countdownTimer float64
	gameStarted bool
}

func (gs *GameScene) Init(
	app *App,
	globalSettings GlobalTetrisSettings,
	objectiveSettings ObjectiveSettings,
) {
	gs.app = app
	gs.seed = time.Now().UnixNano()
	gs.es = NewTetrisField(gs.seed, globalSettings)

	gs.objectiveSettings = objectiveSettings
	gs.objective = gs.objectiveSettings.Init(gs.es)

	gs.countdownTimer = COUNTDOWN_DURATION_SECS
	gs.gameStarted = false
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
	default:
		if gs.gameStarted {
			gs.objective.HandleAction(act, gs.es)
		}
	}
}

func (gs *GameScene) Update() {
	if gs.gameStarted {
		gs.objective.Update(gs.es)
		return
	}

	gs.countdownTimer -= UPDATE_TICK_RATE_MS/1000.0
	if gs.countdownTimer < 0 {
		gs.gameStarted = true
		gs.es.gameStarted = true
		gs.es.GetRandomPiece()
	}
}

func (gs *GameScene) Draw(sw, sh int, rr Area, lag float64) {
	playingField := rr.Inset(BOARD_WIDTH, BOARD_HEIGHT+4)
	anchorX := playingField.X - 2
	anchorY := playingField.Bottom() - 2

	gs.es.Draw(sw, sh, playingField, lag)
	gs.es.DrawStats(rr, anchorX, anchorY)

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
	}
}
