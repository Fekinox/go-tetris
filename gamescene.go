package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

type GameScene struct {
	app *App
	es *TetrisField

	seed int64
	globalSettings GlobalTetrisSettings
	objectiveSettings ObjectiveSettings
	objective Objective
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
	default:
		gs.objective.HandleAction(act, gs.es)
	}
}

func (gs *GameScene) Update() {
	gs.objective.Update(gs.es)
}

func (gs *GameScene) Draw(sw, sh int, rr Area, lag float64) {
	playingField := rr.Inset(BOARD_WIDTH, BOARD_HEIGHT+4)
	anchorX := playingField.X - 2
	anchorY := playingField.Bottom() - 2

	gs.es.Draw(sw, sh, playingField, lag)
	gs.es.DrawStats(rr, anchorX, anchorY)
}
