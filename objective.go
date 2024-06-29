package main

import "fmt"

type GlobalTetrisSettings struct {
	StartingLevel   int64
	MaxResets       int64
	LockDelay       int64
	BaseGravity     int64
	GravityIncrease int64
}

var DefaultTetrisSettings = GlobalTetrisSettings{
	StartingLevel:   1,
	MaxResets:       MAX_MOVE_RESETS,
	LockDelay:       LOCK_DELAY,
	BaseGravity:     BASE_GRAVITY,
	GravityIncrease: BASE_GRAVITY_INCREASE,
}

type Objective interface {
	Update(es *TetrisField)
	HandleAction(act Action, es *TetrisField)
	GetStats() []Stat
}

type ObjectiveID int8

const (
	LineClear ObjectiveID = iota
	Survival
	Endless
	Cheese
	ScoreAttack
)

type ObjectiveSettings interface {
	Init(es *TetrisField) Objective

	CreateFormFields() []FormField
}

func (gts *GlobalTetrisSettings) CreateFormFields() []FormField {
	return []FormField{
		NewIntegerField(
			"Starting Level",
			gts.StartingLevel,
			func(value int64) {
				gts.StartingLevel = value
			},
			WithMin(1),
			WithMax(20),
		),
		NewIntegerField(
			"Maximum Resets",
			gts.MaxResets,
			func(value int64) {
				gts.MaxResets = value
			},
			WithMin(0),
		),
		NewIntegerField(
			"Lock Delay",
			gts.LockDelay,
			func(value int64) {
				gts.LockDelay = value
			},
			WithMin(0),
		),
		NewIntegerField(
			fmt.Sprintf("Base Gravity (1/%vG)", BASE_GRAVITY_UNIT),
			gts.BaseGravity,
			func(value int64) {
				gts.BaseGravity = value
			},
			WithMin(0),
		),
		NewIntegerField(
			fmt.Sprintf("Gravity Increase (1/%vG)", BASE_GRAVITY_UNIT),
			gts.GravityIncrease,
			func(value int64) {
				gts.GravityIncrease = value
			},
			WithMin(0),
		),
	}
}
