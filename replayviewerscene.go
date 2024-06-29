package main

import (
	"math"

	"github.com/gdamore/tcell/v2"
)

// TODO: replays look deterministic but implementing some basic tests
// might be smart
// TODO: game scene and replay viewer scene have extremely similar logic
// besides the fact that the replay viewer automatically performs some actions.
// so it might be economical to merge the two with dependency injection or
// something
type ReplayViewerScene struct {
	app *App
	es  *TetrisField

	objective Objective

	replayData     ReplayData
	countdownTimer float64
	countdownSpeed float64
	gameStarted    bool

	actionPointer int
}

func (rvs *ReplayViewerScene) Init(
	app *App,
	replayData ReplayData,
) {
	rvs.app = app
	rvs.replayData = replayData
	rvs.es = NewTetrisField(rvs.replayData.Seed, rvs.replayData.TetrisSettings)
	rvs.es.RegisterAudio(app.Audio)

	rvs.objective = rvs.replayData.ObjectiveSettings.Init(rvs.es)

	rvs.countdownTimer = COUNTDOWN_DURATION_SECS
	rvs.countdownSpeed = COUNTDOWN_SPEED
	
	rvs.gameStarted = false
}

func (rvs *ReplayViewerScene) HandleEvent(ev tcell.Event) {
}

func (rvs *ReplayViewerScene) HandleAction(act Action) {
	switch act {
	case Quit:
		rvs.app.OpenMenuScene()
	case Reset:
		rvs.es.HandleReset(rvs.replayData.Seed)
		rvs.objective = rvs.replayData.ObjectiveSettings.Init(rvs.es)

		rvs.countdownTimer = COUNTDOWN_DURATION_SECS
		rvs.countdownSpeed = RESET_COUNTDOWN_SPEED
		rvs.gameStarted = false

		rvs.actionPointer = 0
	}
}

func (rvs *ReplayViewerScene) Update() {
	if !rvs.gameStarted {
		rvs.countdownTimer -= (UPDATE_TICK_RATE_MS / 1000.0) * rvs.countdownSpeed
		if rvs.countdownTimer < 0 {
			rvs.gameStarted = true
			rvs.es.gameStarted = true
			rvs.es.GetRandomPiece()
		}

		return
	}

	for rvs.actionPointer < len(rvs.replayData.Actions) &&
		rvs.es.frameCount == rvs.replayData.Actions[rvs.actionPointer].Frame {
		act := rvs.replayData.Actions[rvs.actionPointer]
		rvs.objective.HandleAction(act.Action, rvs.es)
		rvs.actionPointer++
	}

	rvs.objective.Update(rvs.es)
}

func (rvs *ReplayViewerScene) Draw(sw, sh int, rr Area, lag float64) {
	playingField := rr.Inset(BOARD_WIDTH, BOARD_HEIGHT+4)
	anchorX := playingField.X - 2
	anchorY := playingField.Bottom() - 2

	rvs.es.Draw(sw, sh, playingField, lag)

	yOffset := 0
	for _, stat := range rvs.objective.GetStats() {
		strings := stat.Compute()
		SetStringArray(
			anchorX,
			anchorY+yOffset-len(strings),
			defStyle,
			true,
			strings...,
		)
		yOffset -= len(strings) + 1
	}

	if !rvs.gameStarted {
		textAnchorX := playingField.X + BOARD_WIDTH/2
		textAnchorY := playingField.Y + 4
		var theText string
		if rvs.countdownTimer > 3.0 {
			theText = "3..."
		} else if rvs.countdownTimer > 2.0 {
			theText = "2..."
		} else if rvs.countdownTimer > 1.0 {
			theText = "1..."
		} else {
			theText = "GO!!"
		}

		SetCenteredString(textAnchorX, textAnchorY, theText, defStyle)
		rvs.DrawProgressBar(
			textAnchorX, textAnchorY+1,
			rvs.countdownTimer-math.Floor(rvs.countdownTimer),
		)
	}
}

func (rvs *ReplayViewerScene) DrawProgressBar(
	anchorX, anchorY int,
	value float64,
) {
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

func (rvs *ReplayViewerScene) Cleanup() {
}
