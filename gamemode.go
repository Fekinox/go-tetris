package main

type GlobalTetrisSettings struct {
	StartingLevel int64
	MaxResets int
	LockDelay int
	BaseGravity int64
	GravityIncrease int64
}

type Objective interface {
	Update(es *TetrisField)
	HandleAction(act Action, es *TetrisField)
}

type ObjectiveID int

type ObjectiveSettings interface {
	Init(es *TetrisField) Objective
}
