package main

import "fmt"

const MAX_CHEESE_GARBAGE_LINES = 10

type CheeseSettings struct {
	Garbage int64
	Endless bool
}

type CheeseObjective struct {
	CurrentGarbage int64
	GarbageLimit int64
	Endless bool

	stats []Stat
}

func (cs *CheeseSettings) Init(es *TetrisField) Objective {
	co := &CheeseObjective{
		CurrentGarbage: 0,
		GarbageLimit: cs.Garbage,
		Endless: cs.Endless,

		stats: []Stat {
			CreateElapsedTimeStat(es),
			CreateLinesStat(es),
			CreatePiecesStat(es),
		},
	}

	if co.Endless {
		for i := int64(0); i < MAX_CHEESE_GARBAGE_LINES; i++ {
			es.AddGarbage(1)
		}
	} else {
		for i := int64(0); i < min(co.GarbageLimit - co.CurrentGarbage, MAX_CHEESE_GARBAGE_LINES); i++ {
			es.AddGarbage(1)
		}
	}

	es.AddLineClearHandler(func(garbage, nonGarbage int) {
		co.OnLineClear(garbage, es)
	})

	co.stats = append(co.stats, Stat{
		Compute: func() []string {
			var remGarbage string
			if co.Endless {
				remGarbage = "inf"
			} else {
				remGarbage = fmt.Sprint(co.GarbageLimit)
			}
			return []string{
				"GARBAGE",
				fmt.Sprintf("%v/%v", co.CurrentGarbage, remGarbage),
			}
		},
	})

	return co
}

func (co *CheeseObjective) GetStats() []Stat {
	return co.stats
}

func (co *CheeseObjective) Update(es *TetrisField) {
	if es.gameOver {
		return
	}

	es.Update()
}

func (co *CheeseObjective) HandleAction(act Action, es *TetrisField) {
	es.HandleAction(act)
}

func (co *CheeseObjective) OnLineClear(garbage int, es *TetrisField) {
	for i := 0; i < garbage; i++ {
		co.CurrentGarbage += 1
		if !co.Endless && co.CurrentGarbage == co.GarbageLimit {
			es.ObjectiveComplete("Cleared all garbage")
		}
		if co.Endless || co.GarbageLimit - co.CurrentGarbage >= MAX_CHEESE_GARBAGE_LINES {
			es.QueueGarbage(1)
		}
	}
}

func (cs *CheeseSettings) CreateFormFields() []FormField {
	return []FormField{
		NewIntegerField(
			"Garbage",
			cs.Garbage,
			func(value int64) {
				cs.Garbage = value
			},
			WithMin(1),
		),
		NewBooleanField(
			"Endless",
			cs.Endless,
			func(value bool) {
				cs.Endless = value
			},
		),
	}
}
