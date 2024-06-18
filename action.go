package main

type Action int

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

type ReplayAction struct {
	Action Action
	Frame int64
}
