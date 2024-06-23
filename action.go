package main

type Action int8

const (
	MoveUp Action = iota
	MoveDown
	MoveLeft
	MoveRight
	HardDrop
	RotateCW
	RotateCCW
	SwapHoldPiece
	ToggleSuper
	Quit
	Reset
	Pause
	MenuConfirm
)

var ActionNames = []string{
	"MoveUp",
	"MoveDown",
	"MoveLeft",
	"MoveRight",
	"HardDrop",
	"RotateCW",
	"RotateCCW",
	"SwapHoldPiece",
	"ToggleSuper",
	"Quit",
	"Reset",
	"Pause",
	"MenuConfirm",
}

type ReplayAction struct {
	Action Action
	Frame  int64
}

func (a Action) ToString() string {
	return ActionNames[a]
}
