package main

const MAX_CHEESE_GARBAGE_LINES = 10

type CheeseSettings struct {
	Garbage int	
}

func (cs *CheeseSettings) Init(es *TetrisField) Objective {
	co := &CheeseObjective{
		CurrentGarbage: cs.Garbage,
	}

	for i := 0; i < co.CurrentGarbage && i < MAX_CHEESE_GARBAGE_LINES; i++ {
		es.AddGarbage(1)
	}

	es.AddLineClearHandler(func(garbage, nonGarbage int) {
		co.OnLineClear(garbage, es)
	})

	return co
}

type CheeseObjective struct {
	CurrentGarbage int
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
		co.CurrentGarbage -= 1
		if co.CurrentGarbage == 0 {
			es.gameOver = true
		}
		if co.CurrentGarbage >= MAX_CHEESE_GARBAGE_LINES {
			es.QueueGarbage(1)
		}
	}
}
