package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type PreGameScene struct {
	app *App

	objectiveID       ObjectiveID
	tetrisSettings    GlobalTetrisSettings
	objectiveSettings ObjectiveSettings

	tetrisFormFields    []FormField
	objectiveFormFields []FormField

	menuFocus    int
	editingField bool
}

func (pgs *PreGameScene) Init(
	app *App,
	objectiveID ObjectiveID,
	tSettings GlobalTetrisSettings,
	oSettings ObjectiveSettings,
) {
	pgs.app = app
	pgs.objectiveID = objectiveID
	pgs.tetrisSettings = tSettings
	pgs.objectiveSettings = oSettings

	pgs.tetrisFormFields = pgs.tetrisSettings.CreateFormFields()
	pgs.objectiveFormFields = pgs.objectiveSettings.CreateFormFields()

}

func (pgs *PreGameScene) HandleEvent(ev tcell.Event) {
	if pgs.editingField {
		idx := pgs.menuFocus - 1
		if idx < len(pgs.tetrisFormFields) {
			pgs.tetrisFormFields[idx].Field.HandleInput(ev)
		} else {
			idx -= len(pgs.tetrisFormFields)
			pgs.objectiveFormFields[idx].Field.HandleInput(ev)
		}
	}
}

func (pgs *PreGameScene) HandleAction(act Action) {
	switch act {
	case MoveUp:
		pgs.editingField = false
		pgs.menuFocus = max(
			0,
			pgs.menuFocus-1,
		)
	case MoveDown:
		pgs.editingField = false
		pgs.menuFocus = min(
			len(pgs.tetrisFormFields)+len(pgs.objectiveFormFields),
			pgs.menuFocus+1,
		)
	case MenuConfirm:
		if pgs.menuFocus == 0 {
			pgs.app.OpenGameScene(
				pgs.tetrisSettings,
				pgs.objectiveID,
				pgs.objectiveSettings,
			)
		} else {
			// If the current field is a boolean field, toggle its value
			var field EditableField
			idx := pgs.menuFocus - 1
			if idx < len(pgs.tetrisFormFields) {
				field = pgs.tetrisFormFields[idx].Field
			} else {
				idx -= len(pgs.tetrisFormFields)
				field = pgs.objectiveFormFields[idx].Field
			}

			if field, ok := field.(*BooleanField); ok {
				field.SetValue(!field.Value)
			} else {
				pgs.editingField = !pgs.editingField
			}
		}
	case Quit:
		pgs.app.OpenMenuScene()
	}
}

func (pgs *PreGameScene) Update() {
}

func (pgs *PreGameScene) Draw(sw, sh int, rr Area, lag float64) {
	yPosition := pgs.menuFocus * 2
	if pgs.menuFocus > 0 {
		yPosition += 2
	}
	if pgs.menuFocus > len(pgs.tetrisFormFields) {
		yPosition += 2
	}

	focusStyle := defStyle.Reverse(true)

	// Draw focus marker
	Screen.SetContent(
		rr.X,
		rr.Y+yPosition,
		'*',
		nil, defStyle)

	// Draw confirm button
	if yPosition == 0 {
		SetString(
			rr.X+2,
			rr.Y+0,
			"Start Game",
			focusStyle)
	} else {
		SetString(
			rr.X+2,
			rr.Y+0,
			"Start Game",
			defStyle)
	}

	// Draw tetris settings
	SetString(
		rr.X+2,
		rr.Y+2,
		"Tetris Settings",
		defStyle)

	for i, opt := range pgs.tetrisFormFields {
		position := 4 + 2*i
		style := defStyle
		if position == yPosition && !pgs.editingField {
			style = focusStyle
		}
		SetString(
			rr.X+2,
			rr.Y+position,
			opt.Name,
			style,
		)

		opt.Field.Draw(
			rr.X+3+runewidth.StringWidth(opt.Name),
			rr.Y+position,
			pgs.editingField && position == yPosition,
		)
	}

	// Draw objective settings
	if len(pgs.objectiveFormFields) > 0 {
		SetString(
			rr.X+2,
			rr.Y+2*(2+len(pgs.tetrisFormFields)),
			"Objective Settings",
			defStyle)

		for i, opt := range pgs.objectiveFormFields {
			position := 4 + 2*(1+len(pgs.tetrisFormFields)+i)
			style := defStyle
			if position == yPosition {
				style = focusStyle
			}
			SetString(
				rr.X+2,
				rr.Y+position,
				opt.Name,
				style,
			)

			opt.Field.Draw(
				rr.X+3+runewidth.StringWidth(opt.Name),
				rr.Y+position,
				pgs.editingField && position == yPosition,
			)
		}
	}
}

func (pgs *PreGameScene) Cleanup() {
}
