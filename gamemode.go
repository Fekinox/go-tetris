package main

type GlobalTetrisSettings struct {
	StartingLevel   int64
	MaxResets       int64
	LockDelay       int64
	BaseGravity     int64
	GravityIncrease int64
}

type Objective interface {
	Update(es *TetrisField)
	HandleAction(act Action, es *TetrisField)
}

type ObjectiveID int8

const (
	LineClear ObjectiveID = iota
	Survival
	Endless
	Cheese
)

type ObjectiveSettings interface {
	Init(es *TetrisField) Objective
}
